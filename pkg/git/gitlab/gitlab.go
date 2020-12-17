package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

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
	// TODO
	var webhook git.Webhook
	if err := Validate(integrationConfig.Status.Secrets, header.Get("x-gitlab-token")); err != nil {
		return webhook, err
	}

	var eventType git.EventType

	var err error

	eventFromHeader := header.Get("x-gitlab-event")
	if eventFromHeader == "Merge Request Hook" {
		eventType = git.EventTypePullRequest
	} else if eventFromHeader == "Push Hook" || eventFromHeader == "Tag Push Hook" {
		eventType = git.EventTypePush
	}

	if eventType == git.EventTypePullRequest {
		var data MergeRequestWebhook

		if err = json.Unmarshal(jsonString, &data); err != nil {
			return git.Webhook{}, err
		}
		sender := git.Sender{Name: data.Sender.ID}
		base := git.Base{Ref: data.ObjectAttribute.BaseRef}
		head := git.Head{Ref: data.ObjectAttribute.HeadRef, Sha: data.ObjectAttribute.LastCommit.Sha}
		repo := git.Repository{Name: data.Repo.Name, URL: data.Repo.Htmlurl}
		pullRequest := git.PullRequest{ID: data.ObjectAttribute.ID, Title: data.ObjectAttribute.Title, Sender: sender, URL: data.Repo.Htmlurl, Base: base, Head: head}
		webhook = git.Webhook{EventType: git.EventType(eventType), Repo: repo, PullRequest: &pullRequest}

	} else if eventType == git.EventTypePush {
		var data PushWebhook

		if err = json.Unmarshal(jsonString, &data); err != nil {
			return git.Webhook{}, err
		}
		repo := git.Repository{Name: data.Repo.Name, URL: data.Repo.Htmlurl}
		push := git.Push{Pusher: data.User, Ref: data.Ref, Sha: data.Sha}
		webhook = git.Webhook{EventType: git.EventType(eventType), Repo: repo, Push: &push}

	} else {
		return webhook, fmt.Errorf("event %s is not supported", eventType)
	}
	return webhook, nil
}

func (c *Client) RegisterWebhook(integrationConfig *cicdv1.IntegrationConfig, Url string, client *client.Client) error {
	// TODO
	var registrationBody RegistrationWebhookBody
	EncodedRepoPath := url.QueryEscape(integrationConfig.Spec.Git.Repository)
	apiURL := integrationConfig.Spec.Git.GetApiUrl() + "/api/v4/projects/" + EncodedRepoPath + "/hooks"
	var httpClient = &http.Client{}

	//enable hooks from every events
	registrationBody.ConfidentialIssueEvents = true
	registrationBody.ConfidentialNoteEvents = true
	registrationBody.DeploymentEvents = true
	registrationBody.IssueEvents = true
	registrationBody.JobEvents = true
	registrationBody.MergeRequestEvents = true
	registrationBody.PipeLineEvents = true
	registrationBody.PushEvents = true
	registrationBody.TagPushEvents = true
	registrationBody.WikiPageEvents = true
	registrationBody.URL = Url
	registrationBody.ID = EncodedRepoPath

	jsonBytes, _ := json.Marshal(registrationBody)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	token, err := integrationConfig.GetToken(*client)

	if err != nil {
		return err
	}

	req.Header.Add("PRIVATE-TOKEN", token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("error setting webhook, code %d, msg %s", resp.StatusCode, string(body))
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}

	return nil
}

func (c *Client) SetCommitStatus(integrationJob *cicdv1.IntegrationJob, integrationConfig *cicdv1.IntegrationConfig, context string, state git.CommitStatusState, description, targetUrl string, client *client.Client) error {
	// TODO
	var commitStatusBody CommitStatusBody
	var httpClient = &http.Client{}
	var urlEncodePath = url.QueryEscape(integrationConfig.Spec.Git.Repository)
	var sha string
	if integrationJob.Spec.Refs.Pull == nil {
		sha = integrationJob.Spec.Refs.Base.Sha
	} else {
		sha = integrationJob.Spec.Refs.Pull.Sha
	}
	apiUrl := integrationConfig.Spec.Git.GetApiUrl() + "/api/v4/projects/" + urlEncodePath + "/statuses/" + sha
	commitStatusBody.State = string(state)
	commitStatusBody.TargetURL = targetUrl
	commitStatusBody.Description = description
	commitStatusBody.Context = context

	jsonBytes, err := json.Marshal(commitStatusBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil
	}

	token, err := integrationConfig.GetToken(*client)
	if err != nil {
		return err
	}
	req.Header.Add("PRIVATE-TOKEN", token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("error setting commit status, code %d, msg %s", resp.StatusCode, string(body))
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}
	return nil
}

func Validate(secret, headerToken string) error {
	if secret != headerToken {
		return fmt.Errorf("invalid request : X-Gitlab-Token does not match secret")
	}
	return nil
}
