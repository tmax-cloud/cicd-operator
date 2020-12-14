package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
}

func (c *Client) ParseWebhook(header http.Header, jsonString []byte) (git.Webhook, error) {
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
	var registrationBody RegistrationWebhookBody
	var registrationConfig RegistrationWebhookBodyConfig
	var apiUrl string = integrationConfig.Spec.Git.GetApiUrl() + "/repos/" + integrationConfig.Spec.Git.Repository + "/hooks"
	var httpClient = &http.Client{}

	registrationBody.Name = "web"
	registrationBody.Active = true
	registrationBody.Events = []string{"*"}
	registrationConfig.Url = url
	registrationConfig.ContentType = "json"
	registrationConfig.InsecureSsl = "0"
	registrationConfig.Secret = integrationConfig.Status.Secrets

	registrationBody.Config = registrationConfig
	jsonBytes, err := json.Marshal(registrationBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	token, err := integrationConfig.GetToken(*client)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "token "+token)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

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

func (c *Client) SetCommitStatus(gitConfig *cicdv1.GitConfig, context string, state git.CommitStatusState, token, description, targetUrl string) error {

	var commitStatusBody CommitStatusBody
	var httpClient = &http.Client{}
	apiUrl := gitConfig.GetApiUrl() + "/repos/" + gitConfig.Repository + "/statuses/" + sha
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
	req.Header.Add("Authorization", "token "+token)
	req.Header.Add("Accept", "application/vnd.github.v3+json")

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
