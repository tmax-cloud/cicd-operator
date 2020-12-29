package github

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
}

type CommitStatusBody struct {
	State       string `json:"state"`
	TargetURL   string `json:"target_url"`
	Description string `json:"description"`
	Context     string `json:"context"`
}

func (c *Client) ParseWebhook(integrationConfig *cicdv1.IntegrationConfig, header http.Header, jsonString []byte) (git.Webhook, error) {
	var webhook git.Webhook
	var signature = strings.Replace(header.Get("x-hub-signature"), "sha1=", "", 1)
	if err := Validate(integrationConfig.Status.Secrets, signature, jsonString); err != nil {
		return webhook, err
	}
	var eventType = git.EventType(header.Get("x-github-event"))

	var err error
	if eventType == git.EventTypePullRequest {
		var data PullRequestWebhook

		if err = json.Unmarshal(jsonString, &data); err != nil {
			return git.Webhook{}, err
		}
		sender := git.Sender{Name: data.Sender.ID}
		base := git.Base{Ref: data.PullRequest.Base.Ref}
		head := git.Head{Ref: data.PullRequest.Head.Ref, Sha: data.PullRequest.Head.Sha}
		repo := git.Repository{Name: data.Repo.Name, URL: data.Repo.Htmlurl}
		pullRequest := git.PullRequest{ID: data.Number, Title: data.PullRequest.Title, Sender: sender, URL: data.Repo.Htmlurl, Base: base, Head: head, State: git.PullRequestState(data.PullRequest.State), Action: git.PullRequestAction(data.Action)}
		webhook = git.Webhook{EventType: eventType, Repo: repo, PullRequest: &pullRequest}

	} else if eventType == git.EventTypePush {
		var data PushWebhook

		if err = json.Unmarshal(jsonString, &data); err != nil {
			return git.Webhook{}, err
		}
		repo := git.Repository{Name: data.Repo.Name, URL: data.Repo.Htmlurl}
		var sha string
		if strings.Contains(data.Ref, "refs/tags") {
			sha = data.Sha4Tag
		} else {
			sha = data.Sha4Push
		}
		push := git.Push{Pusher: data.Pusher.Name, Ref: data.Ref, Sha: sha}
		webhook = git.Webhook{EventType: eventType, Repo: repo, Push: &push}

	} else {
		return webhook, nil
	}
	return webhook, nil
}

func (c *Client) ListWebhook(integrationConfig *cicdv1.IntegrationConfig, client client.Client) ([]git.WebhookEntry, error) {
	var apiUrl = integrationConfig.Spec.Git.GetApiUrl() + "/repos/" + integrationConfig.Spec.Git.Repository + "/hooks"

	token, err := integrationConfig.GetToken(client)
	if err != nil {
		return nil, err
	}
	header := map[string]string{"Authorization": "token " + token}
	data, _, err := git.RequestHttp(http.MethodGet, apiUrl, header, nil)
	if err != nil {
		return nil, err
	}

	var entries []WebhookEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	var result []git.WebhookEntry
	for _, e := range entries {
		result = append(result, git.WebhookEntry{Id: e.Id, Url: e.Config.Url})
	}

	return result, nil
}

func (c *Client) RegisterWebhook(integrationConfig *cicdv1.IntegrationConfig, url string, client client.Client) error {
	var registrationBody RegistrationWebhookBody
	var registrationConfig RegistrationWebhookBodyConfig
	var apiUrl = integrationConfig.Spec.Git.GetApiUrl() + "/repos/" + integrationConfig.Spec.Git.Repository + "/hooks"

	isUnique, err := CheckUniqueness(integrationConfig, url, client)
	if err != nil {
		return err
	}

	if !isUnique {
		return fmt.Errorf("same webhook has already registered")
	}

	registrationBody.Name = "web"
	registrationBody.Active = true
	registrationBody.Events = []string{"*"}
	registrationConfig.Url = url
	registrationConfig.ContentType = "json"
	registrationConfig.InsecureSsl = "0"
	registrationConfig.Secret = integrationConfig.Status.Secrets

	registrationBody.Config = registrationConfig

	token, err := integrationConfig.GetToken(client)
	if err != nil {
		return err
	}
	header := map[string]string{"Authorization": "token " + token}
	_, _, err = git.RequestHttp(http.MethodPost, apiUrl, header, registrationBody)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteWebhook(integrationConfig *cicdv1.IntegrationConfig, id int, client client.Client) error {
	var apiUrl = integrationConfig.Spec.Git.GetApiUrl() + "/repos/" + integrationConfig.Spec.Git.Repository + "/hooks/" + strconv.Itoa(id)
	token, err := integrationConfig.GetToken(client)
	if err != nil {
		return err
	}
	header := map[string]string{"Authorization": "token " + token}
	_, _, err = git.RequestHttp(http.MethodDelete, apiUrl, header, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SetCommitStatus(integrationJob *cicdv1.IntegrationJob, integrationConfig *cicdv1.IntegrationConfig, context string, state git.CommitStatusState, description, targetUrl string, client client.Client) error {
	var commitStatusBody CommitStatusBody
	var sha string
	if integrationJob.Spec.Refs.Pull == nil {
		sha = integrationJob.Spec.Refs.Base.Sha
	} else {
		sha = integrationJob.Spec.Refs.Pull.Sha
	}
	apiUrl := integrationConfig.Spec.Git.GetApiUrl() + "/repos/" + integrationJob.Spec.Refs.Repository + "/statuses/" + sha

	commitStatusBody.State = string(state)
	commitStatusBody.TargetURL = targetUrl
	commitStatusBody.Description = description
	commitStatusBody.Context = context

	token, err := integrationConfig.GetToken(client)
	if err != nil {
		return err
	}
	header := map[string]string{
		"Authorization": "token " + token,
		"Accept":        "application/vnd.github.v3+json",
	}
	_, _, err = git.RequestHttp(http.MethodPost, apiUrl, header, commitStatusBody)
	if err != nil {
		return err
	}

	return nil
}

func IsValidPayload(secret, headerHash string, payload []byte) bool {
	hash := HashPayload(secret, payload)
	return hmac.Equal(
		[]byte(hash),
		[]byte(headerHash),
	)
}

func HashPayload(secret string, payloadBody []byte) string {
	hm := hmac.New(sha1.New, []byte(secret))
	_, err := hm.Write(payloadBody)
	sum := hm.Sum(nil)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", sum)
}

func Validate(secret, headerHash string, payload []byte) error {
	if !IsValidPayload(secret, headerHash, payload) {
		return fmt.Errorf("invalid request : X-Hub-Signature does not match secret")
	}
	return nil
}

func CheckUniqueness(integrationConfig *cicdv1.IntegrationConfig, url string, client client.Client) (bool, error) {
	var apiUrl = integrationConfig.Spec.Git.GetApiUrl() + "/repos/" + integrationConfig.Spec.Git.Repository + "/hooks"

	token, err := integrationConfig.GetToken(client)
	if err != nil {
		return false, err
	}
	header := map[string]string{"Authorization": "token " + token}
	data, _, err := git.RequestHttp(http.MethodGet, apiUrl, header, nil)
	if err != nil {
		return false, err
	}

	var entries []WebhookEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return false, err
	}

	for _, e := range entries {
		if url == e.Config.Url {
			return false, nil
		}
	}

	return true, nil
}
