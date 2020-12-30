package gitlab

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
		repo := git.Repository{Name: data.Project.Name, URL: data.Project.WebUrl}
		action := git.PullRequestAction(data.ObjectAttribute.Action)
		switch string(action) {
		case "close":
			action = git.PullRequestActionClose
		case "open":
			action = git.PullRequestActionOpen
		case "reopen":
			action = git.PullRequestActionReOpen
		case "update":
			action = git.PullRequestActionSynchronize
		}
		state := git.PullRequestState(data.ObjectAttribute.State)
		switch string(state) {
		case "opened":
			state = git.PullRequestStateOpen
		case "closed":
			state = git.PullRequestStateClosed
		}
		pullRequest := git.PullRequest{ID: data.ObjectAttribute.ID, Title: data.ObjectAttribute.Title, Sender: sender, URL: data.Project.WebUrl, Base: base, Head: head, State: state, Action: action}
		webhook = git.Webhook{EventType: eventType, Repo: repo, PullRequest: &pullRequest}

	} else if eventType == git.EventTypePush {
		var data PushWebhook

		if err = json.Unmarshal(jsonString, &data); err != nil {
			return git.Webhook{}, err
		}
		repo := git.Repository{Name: data.Repo.Name, URL: data.Repo.Htmlurl}
		push := git.Push{Pusher: data.User, Ref: data.Ref, Sha: data.Sha}
		webhook = git.Webhook{EventType: eventType, Repo: repo, Push: &push}

	} else {
		return webhook, nil
	}
	return webhook, nil
}

func (c *Client) ListWebhook(integrationConfig *cicdv1.IntegrationConfig, client client.Client) ([]git.WebhookEntry, error) {
	encodedRepoPath := url.QueryEscape(integrationConfig.Spec.Git.Repository)
	apiURL := integrationConfig.Spec.Git.GetApiUrl() + "/api/v4/projects/" + encodedRepoPath + "/hooks"

	token, err := integrationConfig.GetToken(client)
	if err != nil {
		return nil, err
	}

	header := map[string]string{
		"PRIVATE-TOKEN": token,
		"Content-Type":  "application/json",
	}
	data, _, err := git.RequestHttp(http.MethodGet, apiURL, header, nil)
	if err != nil {
		return nil, err
	}

	var entries []WebhookEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	var result []git.WebhookEntry
	for _, e := range entries {
		result = append(result, git.WebhookEntry{Id: e.Id, Url: e.Url})
	}

	return result, nil
}

func (c *Client) RegisterWebhook(integrationConfig *cicdv1.IntegrationConfig, Url string, client client.Client) error {
	var registrationBody RegistrationWebhookBody
	EncodedRepoPath := url.QueryEscape(integrationConfig.Spec.Git.Repository)
	apiURL := integrationConfig.Spec.Git.GetApiUrl() + "/api/v4/projects/" + EncodedRepoPath + "/hooks"

	//enable hooks from every events
	registrationBody.EnableSSLVerification = false
	registrationBody.ConfidentialIssueEvents = true
	registrationBody.ConfidentialNoteEvents = true
	registrationBody.DeploymentEvents = true
	registrationBody.IssueEvents = true
	registrationBody.JobEvents = true
	registrationBody.MergeRequestEvents = true
	registrationBody.NoteEvents = true
	registrationBody.PipeLineEvents = true
	registrationBody.PushEvents = true
	registrationBody.TagPushEvents = true
	registrationBody.WikiPageEvents = true
	registrationBody.URL = Url
	registrationBody.ID = EncodedRepoPath
	registrationBody.Token = integrationConfig.Status.Secrets

	token, err := integrationConfig.GetToken(client)
	if err != nil {
		return err
	}
	header := map[string]string{
		"PRIVATE-TOKEN": token,
		"Content-Type":  "application/json",
	}
	_, _, err = git.RequestHttp(http.MethodPost, apiURL, header, registrationBody)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteWebhook(integrationConfig *cicdv1.IntegrationConfig, id int, client client.Client) error {
	encodedRepoPath := url.QueryEscape(integrationConfig.Spec.Git.Repository)
	apiURL := integrationConfig.Spec.Git.GetApiUrl() + "/api/v4/projects/" + encodedRepoPath + "/hooks/" + strconv.Itoa(id)

	token, err := integrationConfig.GetToken(client)
	if err != nil {
		return err
	}

	header := map[string]string{
		"PRIVATE-TOKEN": token,
		"Content-Type":  "application/json",
	}

	_, _, err = git.RequestHttp(http.MethodDelete, apiURL, header, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SetCommitStatus(integrationJob *cicdv1.IntegrationJob, integrationConfig *cicdv1.IntegrationConfig, context string, state git.CommitStatusState, description, targetUrl string, client client.Client) error {
	var commitStatusBody CommitStatusBody
	var urlEncodePath = url.QueryEscape(integrationConfig.Spec.Git.Repository)
	var sha string
	if integrationJob.Spec.Refs.Pull == nil {
		sha = integrationJob.Spec.Refs.Base.Sha
	} else {
		sha = integrationJob.Spec.Refs.Pull.Sha
	}
	apiUrl := integrationConfig.Spec.Git.GetApiUrl() + "/api/v4/projects/" + urlEncodePath + "/statuses/" + sha
	switch cicdv1.CommitStatusState(state) {
	case cicdv1.CommitStatusStatePending:
		commitStatusBody.State = "running"
	case cicdv1.CommitStatusStateFailure, cicdv1.CommitStatusStateError:
		commitStatusBody.State = "failed"
	default:
		commitStatusBody.State = string(state)
	}
	commitStatusBody.TargetURL = targetUrl
	commitStatusBody.Description = description
	commitStatusBody.Context = context

	token, err := integrationConfig.GetToken(client)
	if err != nil {
		return err
	}
	header := map[string]string{
		"PRIVATE-TOKEN": token,
		"Content-Type":  "application/json",
	}
	_, _, err = git.RequestHttp(http.MethodPost, apiUrl, header, commitStatusBody)
	// TODO - error from gitlab
	// Cannot transition status via :run from :running
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "cannot transition status via") {
		err = nil
	}
	if err != nil {
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
