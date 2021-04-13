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

	header map[string]string
}

// Init initiates the Client
func (c *Client) Init() error {
	token, err := c.IntegrationConfig.GetToken(c.K8sClient)
	if err != nil {
		return err
	}

	c.header = map[string]string{
		"Accept": "application/vnd.github.v3+json",
	}
	if token != "" {
		c.header["Authorization"] = "token " + token
	}
	return nil
}

// ParseWebhook parses a webhook body for github
func (c *Client) ParseWebhook(header http.Header, jsonString []byte) (*git.Webhook, error) {
	var signature = strings.Replace(header.Get("x-hub-signature"), "sha1=", "", 1)
	if err := Validate(c.IntegrationConfig.Status.Secrets, signature, jsonString); err != nil {
		return nil, err
	}
	eventType := git.EventType(header.Get("x-github-event"))
	switch eventType {
	case git.EventTypePullRequest:
		return c.parsePullRequestWebhook(jsonString)
	case git.EventTypePush:
		return c.parsePushWebhook(jsonString)
	case git.EventTypeIssueComment:
		return c.parseIssueCommentWebhook(jsonString)
	case git.EventTypePullRequestReview:
		return c.parsePullRequestReviewWebhook(jsonString)
	case git.EventTypePullRequestReviewComment:
		return c.parsePullRequestReviewCommentWebhook(jsonString)
	}
	return nil, nil
}

// ListWebhook lists registered webhooks
func (c *Client) ListWebhook() ([]git.WebhookEntry, error) {
	var apiURL = c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/repos/" + c.IntegrationConfig.Spec.Git.Repository + "/hooks"

	data, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
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

	if _, _, err := c.requestHTTP(http.MethodPost, apiURL, registrationBody); err != nil {
		return err
	}

	return nil
}

// DeleteWebhook deletes registered webhook
func (c *Client) DeleteWebhook(id int) error {
	var apiURL = c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/repos/" + c.IntegrationConfig.Spec.Git.Repository + "/hooks/" + strconv.Itoa(id)
	if _, _, err := c.requestHTTP(http.MethodDelete, apiURL, nil); err != nil {
		return err
	}
	return nil
}

// ListCommitStatuses lists commit status of the specific commit
func (c *Client) ListCommitStatuses(ref string) ([]git.CommitStatus, error) {
	apiURL := c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/repos/" + c.IntegrationConfig.Spec.Git.Repository + "/commits/" + ref + "/statuses"

	raw, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	var statuses []CommitStatusResponse
	if err := json.Unmarshal(raw, &statuses); err != nil {
		return nil, err
	}

	// Temp map for filtering duplicated contexts
	tmp := map[string]struct{}{}

	var resp []git.CommitStatus
	for _, s := range statuses {
		_, exist := tmp[s.Context]
		if exist {
			continue
		}
		tmp[s.Context] = struct{}{}
		resp = append(resp, git.CommitStatus{
			Context: s.Context,
			State:   git.CommitStatusState(s.State),
		})
	}

	return resp, nil
}

// SetCommitStatus sets commit status for the specific commit
func (c *Client) SetCommitStatus(integrationJob *cicdv1.IntegrationJob, context string, state git.CommitStatusState, description, targetURL string) error {
	var commitStatusBody CommitStatusRequest
	var sha string
	if integrationJob.Spec.Refs.Pull == nil {
		sha = integrationJob.Spec.Refs.Base.Sha
	} else {
		sha = integrationJob.Spec.Refs.Pull.Sha
	}

	// Don't set commit status if its' sha is a fake
	if sha == git.FakeSha {
		return nil
	}

	apiURL := c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/repos/" + c.IntegrationConfig.Spec.Git.Repository + "/statuses/" + sha

	commitStatusBody.State = string(state)
	commitStatusBody.TargetURL = targetURL
	commitStatusBody.Description = description
	commitStatusBody.Context = context

	if _, _, err := c.requestHTTP(http.MethodPost, apiURL, commitStatusBody); err != nil {
		return err
	}

	return nil
}

// GetUserInfo gets a user's information
func (c *Client) GetUserInfo(userName string) (*git.User, error) {
	// userName is string!
	apiURL := fmt.Sprintf("%s/users/%s", c.IntegrationConfig.Spec.Git.GetAPIUrl(), userName)

	result, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
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

// CanUserWriteToRepo decides if the user has write permission on the repo
func (c *Client) CanUserWriteToRepo(user git.User) (bool, error) {
	// userName is string!
	apiURL := fmt.Sprintf("%s/repos/%s/collaborators/%s/permission", c.IntegrationConfig.Spec.Git.GetAPIUrl(), c.IntegrationConfig.Spec.Git.Repository, user.Name)

	result, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
	if err != nil {
		return false, err
	}

	var permission UserPermission
	if err := json.Unmarshal(result, &permission); err != nil {
		return false, err
	}

	return permission.Permission == "admin" || permission.Permission == "write", nil
}

// RegisterComment registers comment to an issue
func (c *Client) RegisterComment(_ git.IssueType, issueNo int, body string) error {
	apiUrl := fmt.Sprintf("%s/repos/%s/issues/%d/comments", c.IntegrationConfig.Spec.Git.GetAPIUrl(), c.IntegrationConfig.Spec.Git.Repository, issueNo)

	commentBody := &CommentBody{Body: body}
	if _, _, err := c.requestHTTP(http.MethodPost, apiUrl, commentBody); err != nil {
		return err
	}
	return nil
}

// ListPullRequests gets pull request list
func (c *Client) ListPullRequests(onlyOpen bool) ([]git.PullRequest, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/pulls", c.IntegrationConfig.Spec.Git.GetAPIUrl(), c.IntegrationConfig.Spec.Git.Repository)
	if !onlyOpen {
		apiURL += "?state=all"
	}

	raw, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	var prs []PullRequest
	if err := json.Unmarshal(raw, &prs); err != nil {
		return nil, err
	}

	var result []git.PullRequest
	for _, pr := range prs {
		result = append(result, *convertPullRequestToShared(&pr))
	}

	return result, nil
}

// GetPullRequest gets PR given id
func (c *Client) GetPullRequest(id int) (*git.PullRequest, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/pulls/%d", c.IntegrationConfig.Spec.Git.GetAPIUrl(), c.IntegrationConfig.Spec.Git.Repository, id)

	data, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	pr := &PullRequest{}
	if err := json.Unmarshal(data, &pr); err != nil {
		return nil, err
	}

	return convertPullRequestToShared(pr), nil
}

func convertPullRequestToShared(pr *PullRequest) *git.PullRequest {
	var labels []git.IssueLabel
	for _, l := range pr.Labels {
		labels = append(labels, git.IssueLabel{Name: l.Name})
	}

	return &git.PullRequest{
		ID:    pr.Number,
		Title: pr.Title,
		State: git.PullRequestState(pr.State),
		Sender: git.User{
			ID:   pr.User.ID,
			Name: pr.User.Name,
		},
		URL:    pr.URL,
		Base:   git.Base{Ref: pr.Base.Ref},
		Head:   git.Head{Ref: pr.Head.Ref, Sha: pr.Head.Sha},
		Labels: labels,
	}
}

func (c *Client) requestHTTP(method, apiURL string, data interface{}) ([]byte, http.Header, error) {
	return git.RequestHTTP(method, apiURL, c.header, data)
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
