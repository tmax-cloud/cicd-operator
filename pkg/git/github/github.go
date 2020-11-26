package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
)

type Client struct {
}

func (c *Client) ParseWebhook(header http.Header, jsonString []byte) (git.Webhook, error) {
	// TODO
	var eventType = git.EventType(header.Get("X-Github-Event"))
	var webhook git.Webhook
	var err error
	if eventType == git.GitHubEventTypePullRequest {
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

	} else if eventType == git.GitHubEventTypePush {
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

func (c *Client) RegisterWebhook(gitConfig *cicdv1.GitConfig, url string) error {
	// TODO
	return nil
}

func (c *Client) SetCommitStatus(gitConfig *cicdv1.GitConfig, context string, state git.CommitStatusState, description, targetUrl string) error {
	// TODO
	return nil
}
