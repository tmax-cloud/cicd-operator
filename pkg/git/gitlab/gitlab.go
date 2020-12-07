package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
}

func (c *Client) ParseWebhook(header http.Header, jsonString []byte) (git.Webhook, error) {
	// TODO
	var eventType git.EventType
	var webhook git.Webhook
	var err error
	eventFromHeader := header.Get("X-Gitlab-Event")
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
	registrationBody.URL = integrationConfig.Spec.Git.GetServerAddress()
	registrationBody.ID = EncodedRepoPath

	jsonBytes, _ := json.Marshal(registrationBody)

	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBytes))
	token, _ := integrationConfig.GetToken(*client)

	req.Header.Add("PRIVATE-TOKEN", token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("requesting for webhook registration has failed")
	}
	resp.Body.Close()

	return nil
}

func (c *Client) SetCommitStatus(gitConfig *cicdv1.GitConfig, context string, state git.CommitStatusState, description, targetUrl string) error {
	// TODO
	return nil
}
