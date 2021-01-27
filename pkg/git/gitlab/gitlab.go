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

// Client is a gitlab client struct
type Client struct {
	IntegrationConfig *cicdv1.IntegrationConfig
	K8sClient         client.Client
}

// ParseWebhook parses a webhook body for gitlab
func (c *Client) ParseWebhook(header http.Header, jsonString []byte) (*git.Webhook, error) {
	if err := Validate(c.IntegrationConfig.Status.Secrets, header.Get("x-gitlab-token")); err != nil {
		return nil, err
	}

	eventFromHeader := header.Get("x-gitlab-event")
	if eventFromHeader == "Merge Request Hook" {
		return c.parsePullRequestWebhook(jsonString)
	} else if eventFromHeader == "Push Hook" || eventFromHeader == "Tag Push Hook" {
		return c.parsePushWebhook(jsonString)
	}

	return nil, nil
}

func (c *Client) parsePullRequestWebhook(jsonString []byte) (*git.Webhook, error) {
	var data MergeRequestWebhook

	if err := json.Unmarshal(jsonString, &data); err != nil {
		return nil, err
	}
	sender := git.Sender{Name: data.User.Name, Email: data.User.Email}
	base := git.Base{Ref: data.ObjectAttribute.BaseRef}
	head := git.Head{Ref: data.ObjectAttribute.HeadRef, Sha: data.ObjectAttribute.LastCommit.Sha}
	repo := git.Repository{Name: data.Project.Name, URL: data.Project.WebURL}
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
	pullRequest := git.PullRequest{ID: data.ObjectAttribute.ID, Title: data.ObjectAttribute.Title, Sender: sender, URL: data.Project.WebURL, Base: base, Head: head, State: state, Action: action}
	return &git.Webhook{EventType: git.EventTypePullRequest, Repo: repo, PullRequest: &pullRequest}, nil
}

func (c *Client) parsePushWebhook(jsonString []byte) (*git.Webhook, error) {
	var data PushWebhook

	if err := json.Unmarshal(jsonString, &data); err != nil {
		return nil, err
	}
	repo := git.Repository{Name: data.Project.Name, URL: data.Project.WebURL}
	if strings.HasPrefix(data.Sha, "0000") && strings.HasSuffix(data.Sha, "0000") {
		return nil, nil
	}
	push := git.Push{Sender: git.Sender{Name: data.UserName, ID: data.UserID}, Ref: data.Ref, Sha: data.Sha}

	// Get sender email
	userInfo, err := c.GetUserInfo(strconv.Itoa(data.UserID))
	if err == nil {
		push.Sender.Email = userInfo.Email
	}

	return &git.Webhook{EventType: git.EventTypePush, Repo: repo, Push: &push}, nil
}

// ListWebhook lists registered webhooks
func (c *Client) ListWebhook() ([]git.WebhookEntry, error) {
	encodedRepoPath := url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository)
	apiURL := c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/api/v4/projects/" + encodedRepoPath + "/hooks"

	token, err := c.IntegrationConfig.GetToken(c.K8sClient)
	if err != nil {
		return nil, err
	}

	header := map[string]string{
		"PRIVATE-TOKEN": token,
		"Content-Type":  "application/json",
	}
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
		result = append(result, git.WebhookEntry{ID: e.ID, URL: e.URL})
	}

	return result, nil
}

// RegisterWebhook registers our webhook server to the remote git server
func (c *Client) RegisterWebhook(uri string) error {
	var registrationBody RegistrationWebhookBody
	EncodedRepoPath := url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository)
	apiURL := c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/api/v4/projects/" + EncodedRepoPath + "/hooks"

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
	registrationBody.URL = uri
	registrationBody.ID = EncodedRepoPath
	registrationBody.Token = c.IntegrationConfig.Status.Secrets

	token, err := c.IntegrationConfig.GetToken(c.K8sClient)
	if err != nil {
		return err
	}
	header := map[string]string{
		"PRIVATE-TOKEN": token,
		"Content-Type":  "application/json",
	}
	_, _, err = git.RequestHTTP(http.MethodPost, apiURL, header, registrationBody)
	if err != nil {
		return err
	}

	return nil
}

// DeleteWebhook deletes registered webhook
func (c *Client) DeleteWebhook(id int) error {
	encodedRepoPath := url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository)
	apiURL := c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/api/v4/projects/" + encodedRepoPath + "/hooks/" + strconv.Itoa(id)

	token, err := c.IntegrationConfig.GetToken(c.K8sClient)
	if err != nil {
		return err
	}

	header := map[string]string{
		"PRIVATE-TOKEN": token,
		"Content-Type":  "application/json",
	}

	_, _, err = git.RequestHTTP(http.MethodDelete, apiURL, header, nil)
	if err != nil {
		return err
	}

	return nil
}

// SetCommitStatus sets commit status for the specific commit
func (c *Client) SetCommitStatus(integrationJob *cicdv1.IntegrationJob, context string, state git.CommitStatusState, description, targetURL string) error {
	var commitStatusBody CommitStatusBody
	var urlEncodePath = url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository)
	var sha string
	if integrationJob.Spec.Refs.Pull == nil {
		sha = integrationJob.Spec.Refs.Base.Sha
	} else {
		sha = integrationJob.Spec.Refs.Pull.Sha
	}
	apiURL := c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/api/v4/projects/" + urlEncodePath + "/statuses/" + sha
	switch cicdv1.CommitStatusState(state) {
	case cicdv1.CommitStatusStatePending:
		commitStatusBody.State = "running"
	case cicdv1.CommitStatusStateFailure, cicdv1.CommitStatusStateError:
		commitStatusBody.State = "failed"
	default:
		commitStatusBody.State = string(state)
	}
	commitStatusBody.TargetURL = targetURL
	commitStatusBody.Description = description
	commitStatusBody.Context = context

	token, err := c.IntegrationConfig.GetToken(c.K8sClient)
	if err != nil {
		return err
	}
	header := map[string]string{
		"PRIVATE-TOKEN": token,
		"Content-Type":  "application/json",
	}
	_, _, err = git.RequestHTTP(http.MethodPost, apiURL, header, commitStatusBody)
	// Cannot transition status via :run from :running
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "cannot transition status via") {
		err = nil
	}
	if err != nil {
		return err
	}

	return nil
}

// GetUserInfo gets a user's information
func (c *Client) GetUserInfo(userID string) (*git.User, error) {
	// userID is int!
	apiURL := fmt.Sprintf("%s/api/v4/users/%s", c.IntegrationConfig.Spec.Git.GetAPIUrl(), userID)
	token, err := c.IntegrationConfig.GetToken(c.K8sClient)
	if err != nil {
		return nil, err
	}
	header := map[string]string{
		"PRIVATE-TOKEN": token,
		"Content-Type":  "application/json",
	}

	result, _, err := git.RequestHTTP(http.MethodGet, apiURL, header, nil)
	if err != nil {
		return nil, err
	}

	var userInfo UserInfo
	if err := json.Unmarshal(result, &userInfo); err != nil {
		return nil, err
	}

	email := userInfo.PublicEmail
	if email == "" {
		email = userInfo.Email
	}

	return &git.User{
		ID:    userInfo.ID,
		Name:  userInfo.UserName,
		Email: email,
	}, err
}

// Validate validates the webhook payload
func Validate(secret, headerToken string) error {
	if secret != headerToken {
		return fmt.Errorf("invalid request : X-Gitlab-Token does not match secret")
	}
	return nil
}
