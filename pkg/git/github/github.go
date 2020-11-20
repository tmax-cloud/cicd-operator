package github

import (
	"encoding/json"
	"net/http"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
)

type Client struct {
}

func (c *Client) ParseWebhook(header http.Header, jsonString []byte) (git.Webhook, error) {
	// TODO
	var eventType = header.Get("X-Github-Event")
	var webhook git.Webhook
	if eventType == "pull_request" {
		var data PullRequestWebhook
		err := json.Unmarshal(jsonString, &data)
		if err == nil {
			repo := git.Repository{Name: data.Repo.Name, Owner: data.Repo.Owner.ID, Url: data.Repo.Htmlurl, Private: data.Repo.Private}
			pullRequest := git.PullRequest{Title: data.PullRequest.Title, Sender: data.Sender.ID, Url: data.Repo.Htmlurl}
			webhook = git.Webhook{EventType: git.EventType(eventType), Repo: repo, PullRequest: &pullRequest}
		}
	} else if eventType == "push" {
		var data PushWebhook
		err := json.Unmarshal(jsonString, &data)
		if err == nil {
			repo := git.Repository{Name: data.Repo.Name, Owner: data.Repo.Owner.ID, Url: data.Repo.Htmlurl, Private: data.Repo.Private}
			push := git.Push{Pusher: data.Pusher.Name, Ref: data.Ref}
			webhook = git.Webhook{EventType: git.EventType(eventType), Repo: repo, Push: &push}
		}
	}
	return webhook, nil
}

func (c *Client) RegisterWebhook(gitConfig *cicdv1.GitConfig, url string) error {
	// TODO
	return nil
}

func (c *Client) SetCommitStatus(gitConfig *cicdv1.GitConfig, context string, state git.CommitStatusState, description, targetUrl string) error {
	// TODO
	return nil
}
