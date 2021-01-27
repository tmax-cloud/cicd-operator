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

// Client is a gitlab client struct
type Client struct {
	IntegrationConfig *cicdv1.IntegrationConfig
	K8sClient         client.Client
}

// ParseWebhook parses a webhook body for github
func (c *Client) ParseWebhook(header http.Header, jsonString []byte) (*git.Webhook, error) {
	var signature = strings.Replace(header.Get("x-hub-signature"), "sha1=", "", 1)
	if err := Validate(c.IntegrationConfig.Status.Secrets, signature, jsonString); err != nil {
		return nil, err
	}
	eventType := git.EventType(header.Get("x-github-event"))
	if eventType == git.EventTypePullRequest {
		return c.parsePullRequestWebhook(jsonString)
	} else if eventType == git.EventTypePush {
		return c.parsePushWebhook(jsonString)
	}
	return nil, nil
}

func (c *Client) parsePullRequestWebhook(jsonString []byte) (*git.Webhook, error) {
	var data PullRequestWebhook

	if err := json.Unmarshal(jsonString, &data); err != nil {
		return nil, err
	}

	// Get sender email
	sender := git.Sender{Name: data.Sender.Name, ID: data.Sender.ID}
	userInfo, err := c.GetUserInfo(data.Sender.Name)
	if err == nil {
		sender.Email = userInfo.Email
	}

	base := git.Base{Ref: data.PullRequest.Base.Ref}
	head := git.Head{Ref: data.PullRequest.Head.Ref, Sha: data.PullRequest.Head.Sha}
	repo := git.Repository{Name: data.Repo.Name, URL: data.Repo.HTMLURL}
	pullRequest := git.PullRequest{ID: data.Number, Title: data.PullRequest.Title, Sender: sender, URL: data.Repo.HTMLURL, Base: base, Head: head, State: git.PullRequestState(data.PullRequest.State), Action: git.PullRequestAction(data.Action)}
	return &git.Webhook{EventType: git.EventTypePullRequest, Repo: repo, PullRequest: &pullRequest}, nil
}

func (c *Client) parsePushWebhook(jsonString []byte) (*git.Webhook, error) {
	var data PushWebhook

	if err := json.Unmarshal(jsonString, &data); err != nil {
		return nil, err
	}
	repo := git.Repository{Name: data.Repo.Name, URL: data.Repo.HTMLURL}
	if strings.HasPrefix(data.Sha, "0000") && strings.HasSuffix(data.Sha, "0000") {
		return nil, nil
	}
	push := git.Push{Sender: git.Sender{Name: data.Sender.Name, ID: data.Sender.ID}, Ref: data.Ref, Sha: data.Sha}

	// Get sender email
	userInfo, err := c.GetUserInfo(data.Sender.Name)
	if err == nil {
		push.Sender.Email = userInfo.Email
	}

	return &git.Webhook{EventType: git.EventTypePush, Repo: repo, Push: &push}, nil
}

// ListWebhook lists registered webhooks
func (c *Client) ListWebhook() ([]git.WebhookEntry, error) {
	var apiURL = c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/repos/" + c.IntegrationConfig.Spec.Git.Repository + "/hooks"

	token, err := c.IntegrationConfig.GetToken(c.K8sClient)
	if err != nil {
		return nil, err
	}
	header := map[string]string{"Authorization": "token " + token}
	data, _, err := git.RequestHTTP(http.MethodGet, apiURL, header, nil)
	if err != nil {
		return nil, err
	}

	var entries []WebhookEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	var result []git.WebhookEntry
	for _, e := range entries {
		result = append(result, git.WebhookEntry{ID: e.ID, URL: e.Config.URL})
	}

	return result, nil
}

// RegisterWebhook registers our webhook server to the remote git server
func (c *Client) RegisterWebhook(url string) error {
	var registrationBody RegistrationWebhookBody
	var registrationConfig RegistrationWebhookBodyConfig
	var apiURL = c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/repos/" + c.IntegrationConfig.Spec.Git.Repository + "/hooks"

	registrationBody.Name = "web"
	registrationBody.Active = true
	registrationBody.Events = []string{"*"}
	registrationConfig.URL = url
	registrationConfig.ContentType = "json"
	registrationConfig.InsecureSsl = "0"
	registrationConfig.Secret = c.IntegrationConfig.Status.Secrets

	registrationBody.Config = registrationConfig

	token, err := c.IntegrationConfig.GetToken(c.K8sClient)
	if err != nil {
		return err
	}
	header := map[string]string{"Authorization": "token " + token}
	_, _, err = git.RequestHTTP(http.MethodPost, apiURL, header, registrationBody)
	if err != nil {
		return err
	}

	return nil
}

// DeleteWebhook deletes registered webhook
func (c *Client) DeleteWebhook(id int) error {
	var apiURL = c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/repos/" + c.IntegrationConfig.Spec.Git.Repository + "/hooks/" + strconv.Itoa(id)
	token, err := c.IntegrationConfig.GetToken(c.K8sClient)
	if err != nil {
		return err
	}
	header := map[string]string{"Authorization": "token " + token}
	_, _, err = git.RequestHTTP(http.MethodDelete, apiURL, header, nil)
	if err != nil {
		return err
	}

	return nil
}

// SetCommitStatus sets commit status for the specific commit
func (c *Client) SetCommitStatus(integrationJob *cicdv1.IntegrationJob, context string, state git.CommitStatusState, description, targetURL string) error {
	var commitStatusBody CommitStatusBody
	var sha string
	if integrationJob.Spec.Refs.Pull == nil {
		sha = integrationJob.Spec.Refs.Base.Sha
	} else {
		sha = integrationJob.Spec.Refs.Pull.Sha
	}
	apiURL := c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/repos/" + integrationJob.Spec.Refs.Repository + "/statuses/" + sha

	commitStatusBody.State = string(state)
	commitStatusBody.TargetURL = targetURL
	commitStatusBody.Description = description
	commitStatusBody.Context = context

	token, err := c.IntegrationConfig.GetToken(c.K8sClient)
	if err != nil {
		return err
	}
	header := map[string]string{
		"Authorization": "token " + token,
		"Accept":        "application/vnd.github.v3+json",
	}
	_, _, err = git.RequestHTTP(http.MethodPost, apiURL, header, commitStatusBody)
	if err != nil {
		return err
	}

	return nil
}

// GetUserInfo gets a user's information
func (c *Client) GetUserInfo(userName string) (*git.User, error) {
	// userName is string!
	apiURL := fmt.Sprintf("%s/users/%s", c.IntegrationConfig.Spec.Git.GetAPIUrl(), userName)
	token, err := c.IntegrationConfig.GetToken(c.K8sClient)
	if err != nil {
		return nil, err
	}
	header := map[string]string{
		"Authorization": "token " + token,
		"Accept":        "application/vnd.github.v3+json",
	}

	result, _, err := git.RequestHTTP(http.MethodGet, apiURL, header, nil)
	if err != nil {
		return nil, err
	}

	var userInfo UserInfo
	if err := json.Unmarshal(result, &userInfo); err != nil {
		return nil, err
	}

	return &git.User{
		ID:    userInfo.ID,
		Name:  userInfo.UserName,
		Email: userInfo.Email,
	}, nil
}

// IsValidPayload validates the webhook payload
func IsValidPayload(secret, headerHash string, payload []byte) bool {
	hash := HashPayload(secret, payload)
	return hmac.Equal(
		[]byte(hash),
		[]byte(headerHash),
	)
}

// HashPayload hashes the payload
func HashPayload(secret string, payloadBody []byte) string {
	hm := hmac.New(sha1.New, []byte(secret))
	_, err := hm.Write(payloadBody)
	sum := hm.Sum(nil)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", sum)
}

// Validate validates the webhook payload
func Validate(secret, headerHash string, payload []byte) error {
	if !IsValidPayload(secret, headerHash, payload) {
		return fmt.Errorf("invalid request : X-Hub-Signature does not match secret")
	}
	return nil
}
