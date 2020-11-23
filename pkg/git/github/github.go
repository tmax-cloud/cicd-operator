package github

import (
	"encoding/json"
	"net/http"
	"strings"

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
			sender := git.Sender{Name: data.Sender.ID, Link: data.Sender.Link}
			base := git.Base{Ref: data.PullRequest.Base.Ref, Sha: data.PullRequest.Base.Sha}
			head := git.Head{Ref: data.PullRequest.Head.Ref, Sha: data.PullRequest.Head.Sha}
			repo := git.Repository{Name: data.Repo.Name, Owner: data.Repo.Owner.ID, URL: data.Repo.Htmlurl, Private: data.Repo.Private}
			pullRequest := git.PullRequest{ID: data.PullRequest.ID, Title: data.PullRequest.Title, Sender: sender, URL: data.Repo.Htmlurl, Base: base, Head: head}
			webhook = git.Webhook{EventType: git.EventType(eventType), Repo: repo, PullRequest: &pullRequest}
		}
	} else if eventType == "push" {
		var data PushWebhook
		err := json.Unmarshal(jsonString, &data)
		if err == nil {
			repo := git.Repository{Name: data.Repo.Name, Owner: data.Repo.Owner.ID, URL: data.Repo.Htmlurl, Private: data.Repo.Private}
			var sha string
			if strings.Contains(data.Ref, "refs/tags") {
				sha = data.Sha4Tag
			} else {
				sha = data.Sha4Push
			}
			push := git.Push{Pusher: data.Pusher.Name, Ref: data.Ref, Sha: sha}
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
