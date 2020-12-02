package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
}

func (c *Client) ParseWebhook(header http.Header, jsonString []byte) (git.Webhook, error) {
	// TODO
	var eventType = git.EventType(header.Get("X-Github-Event"))
	var webhook git.Webhook
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
		pullRequest := git.PullRequest{ID: data.PullRequest.ID, Title: data.PullRequest.Title, Sender: sender, URL: data.Repo.Htmlurl, Base: base, Head: head}
		webhook = git.Webhook{EventType: git.EventType(eventType), Repo: repo, PullRequest: &pullRequest}

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
		webhook = git.Webhook{EventType: git.EventType(eventType), Repo: repo, Push: &push}

	} else {
		return webhook, fmt.Errorf("event %s is not supported", eventType)
	}
	return webhook, nil
}

func (c *Client) RegisterWebhook(integrationConfig *cicdv1.IntegrationConfig, url string, client *client.Client) error {
	// TODO
	var registrationBody RegistrationWebhookBody
	var registrationConfig RegistrationWebhookBodyConfig
	var apiUrl string = integrationConfig.Spec.Git.GetApiUrl() + "/repos/" + integrationConfig.Spec.Git.Repository + "/hooks"
	var httpClient = &http.Client{}

	registrationBody.Name = "web"
	registrationBody.Active = true
	registrationBody.Events = []string{"pull", "pull_request"}
	registrationConfig.Url = integrationConfig.Spec.Git.GetServerAddress()
	registrationConfig.ContentType = "json"
	registrationConfig.InsecureSsl = "0"

	jsonBytes, _ := json.Marshal(registrationBody)

	req, _ := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonBytes))
	token, _ := integrationConfig.GetToken(*client)
	req.Header.Add("Authorization", "token "+token)
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
