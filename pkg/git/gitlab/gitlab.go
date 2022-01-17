/*
 Copyright 2021 The CI/CD Operator Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

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

	header map[string]string
}

// Init initiates the Client
func (c *Client) Init() error {
	token, err := c.IntegrationConfig.GetToken(c.K8sClient)
	if err != nil {
		return err
	}

	c.header = map[string]string{
		"Content-Type": "application/json",
	}
	if token != "" {
		c.header["PRIVATE-TOKEN"] = token
	}
	return nil
}

// ParseWebhook parses a webhook body for gitlab
func (c *Client) ParseWebhook(header http.Header, jsonString []byte) (*git.Webhook, error) {
	if err := Validate(c.IntegrationConfig.Status.Secrets, header.Get("x-gitlab-token")); err != nil {
		return nil, err
	}

	eventFromHeader := header.Get("x-gitlab-event")
	switch eventFromHeader {
	case "Merge Request Hook":
		return c.parsePullRequestWebhook(jsonString)
	case "Push Hook", "Tag Push Hook":
		return c.parsePushWebhook(jsonString)
	case "Note Hook":
		return c.parseIssueComment(jsonString)
	}

	return nil, nil
}

// ListWebhook lists registered webhooks
func (c *Client) ListWebhook() ([]git.WebhookEntry, error) {
	encodedRepoPath := url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository)
	apiURL := c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/api/v4/projects/" + encodedRepoPath + "/hooks"

	var entries []WebhookEntry
	tlsConfig := c.IntegrationConfig.GetTLSConfig()

	err := git.GetPaginatedRequest(apiURL, tlsConfig, c.header, func() interface{} {
		return &[]WebhookEntry{}
	}, func(i interface{}) {
		entries = append(entries, *i.(*[]WebhookEntry)...)
	})
	if err != nil {
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

	if _, _, err := c.requestHTTP(http.MethodPost, apiURL, registrationBody); err != nil {
		return err
	}

	return nil
}

// DeleteWebhook deletes registered webhook
func (c *Client) DeleteWebhook(id int) error {
	encodedRepoPath := url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository)
	apiURL := c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/api/v4/projects/" + encodedRepoPath + "/hooks/" + strconv.Itoa(id)

	if _, _, err := c.requestHTTP(http.MethodDelete, apiURL, nil); err != nil {
		return err
	}

	return nil
}

// ListCommitStatuses lists commit status of the specific commit
func (c *Client) ListCommitStatuses(ref string) ([]git.CommitStatus, error) {
	var urlEncodePath = url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository)
	apiURL := c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/api/v4/projects/" + urlEncodePath + "/repository/commits/" + ref + "/statuses"

	var statuses []CommitStatusResponse
	tlsConfig := c.IntegrationConfig.GetTLSConfig()

	err := git.GetPaginatedRequest(apiURL, tlsConfig, c.header, func() interface{} {
		return &[]CommitStatusResponse{}
	}, func(i interface{}) {
		statuses = append(statuses, *i.(*[]CommitStatusResponse)...)
	})
	if err != nil {
		return nil, err
	}

	var resp []git.CommitStatus
	for _, s := range statuses {
		state := git.CommitStatusState(s.Status)
		switch s.Status {
		case "running":
			state = "pending"
		case "failed", "canceled":
			state = "failure"
		}
		resp = append(resp, git.CommitStatus{
			Context:     s.Name,
			State:       state,
			Description: s.Description,
			TargetURL:   s.TargetURL,
		})
	}

	return resp, nil
}

// SetCommitStatus sets commit status for the specific commit
func (c *Client) SetCommitStatus(sha string, status git.CommitStatus) error {
	var commitStatusBody CommitStatusRequest
	var urlEncodePath = url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository)

	// Don't set commit status if its' sha is a fake
	if sha == git.FakeSha {
		return nil
	}

	apiURL := c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/api/v4/projects/" + urlEncodePath + "/statuses/" + sha
	switch cicdv1.CommitStatusState(status.State) {
	case cicdv1.CommitStatusStatePending:
		commitStatusBody.State = "running"
	case cicdv1.CommitStatusStateFailure, cicdv1.CommitStatusStateError:
		commitStatusBody.State = "failed"
	default:
		commitStatusBody.State = string(status.State)
	}
	commitStatusBody.TargetURL = status.TargetURL
	commitStatusBody.Description = status.Description
	commitStatusBody.Context = status.Context

	// Cannot transition status via :run from :running
	if _, _, err := c.requestHTTP(http.MethodPost, apiURL, commitStatusBody); err != nil && !strings.Contains(strings.ToLower(err.Error()), "cannot transition status via") {
		return err
	}

	return nil
}

// GetUserInfo gets a user's information
func (c *Client) GetUserInfo(userID string) (*git.User, error) {
	// userID is int!
	apiURL := fmt.Sprintf("%s/api/v4/users/%s", c.IntegrationConfig.Spec.Git.GetAPIUrl(), userID)

	result, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
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

// CanUserWriteToRepo decides if the user has write permission on the repo
func (c *Client) CanUserWriteToRepo(user git.User) (bool, error) {
	// userID is int!
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/members/all/%d", c.IntegrationConfig.Spec.Git.GetAPIUrl(), url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository), user.ID)

	result, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
	if err != nil {
		return false, err
	}

	var permission UserPermission
	if err := json.Unmarshal(result, &permission); err != nil {
		return false, err
	}

	return permission.AccessLevel >= 30, nil
}

// RegisterComment registers comment to an issue
func (c *Client) RegisterComment(issueType git.IssueType, issueNo int, body string) error {
	var t string
	switch issueType {
	case git.IssueTypeIssue:
		t = "issues"
	case git.IssueTypePullRequest:
		t = "merge_requests"
	default:
		return fmt.Errorf("issue type %s is not supported", issueType)
	}

	apiUrl := fmt.Sprintf("%s/api/v4/projects/%s/%s/%d/notes", c.IntegrationConfig.Spec.Git.GetAPIUrl(), url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository), t, issueNo)

	commentBody := &CommentBody{Body: body}
	if _, _, err := c.requestHTTP(http.MethodPost, apiUrl, commentBody); err != nil {
		return err
	}
	return nil
}

// ListComments lists comments of the issue id
// TODO: Consider Gitlab approve
func (c *Client) ListComments(issueNo int) ([]git.IssueComment, error) {
	var comments []git.IssueComment
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d/notes", c.IntegrationConfig.Spec.Git.GetAPIUrl(), url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository), issueNo)

	raw, _, err := c.requestHTTP(http.MethodGet, apiUrl, nil)
	if err != nil {
		return nil, err
	}
	var noteResponses []NoteResponse
	if err := json.Unmarshal(raw, &noteResponses); err != nil {
		return nil, err
	}
	for _, noteResponse := range noteResponses {
		comments = append(comments, git.IssueComment{
			Comment: git.Comment{
				Body:      noteResponse.Body,
				CreatedAt: noteResponse.CreatedAt,
			},
		})
	}
	return comments, nil
}

// ListPullRequests gets pull request list
func (c *Client) ListPullRequests(onlyOpen bool) ([]git.PullRequest, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests?with_merge_status_recheck=true", c.IntegrationConfig.Spec.Git.GetAPIUrl(), url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository))
	if onlyOpen {
		apiURL += "&state=opened"
	}

	var mrs []MergeRequest
	tlsConfig := c.IntegrationConfig.GetTLSConfig()

	err := git.GetPaginatedRequest(apiURL, tlsConfig, c.header, func() interface{} {
		return &[]MergeRequest{}
	}, func(i interface{}) {
		mrs = append(mrs, *i.(*[]MergeRequest)...)
	})
	if err != nil {
		return nil, err
	}

	var result []git.PullRequest
	for _, mr := range mrs {
		result = append(result, git.PullRequest{
			ID:    mr.ID,
			Title: mr.Title,
			State: convertState(mr.State),
			Author: git.User{
				ID:   mr.Author.ID,
				Name: mr.Author.UserName,
			},
			URL:    mr.WebURL,
			Base:   git.Base{Ref: mr.TargetBranch},
			Head:   git.Head{Ref: mr.SourceBranch, Sha: mr.SHA},
			Labels: convertLabel(mr.Labels),
		})
	}

	return result, nil
}

// GetPullRequest gets pull request info
func (c *Client) GetPullRequest(id int) (*git.PullRequest, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d", c.IntegrationConfig.Spec.Git.GetAPIUrl(), url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository), id)

	raw, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	var mr MergeRequest
	if err := json.Unmarshal(raw, &mr); err != nil {
		return nil, err
	}

	// Target Branch
	// TODO - can we delete this logic...? it consumes another API token limit...
	targetBranch, err := c.GetBranch(mr.TargetBranch)
	if err != nil {
		return nil, err
	}

	return &git.PullRequest{
		ID:    mr.ID,
		Title: mr.Title,
		State: convertState(mr.State),
		Author: git.User{
			ID:   mr.Author.ID,
			Name: mr.Author.UserName,
		},
		URL:       mr.WebURL,
		Base:      git.Base{Ref: mr.TargetBranch, Sha: targetBranch.CommitID},
		Head:      git.Head{Ref: mr.SourceBranch, Sha: mr.SHA},
		Labels:    convertLabel(mr.Labels),
		Mergeable: !mr.HasConflicts,
	}, nil
}

// MergePullRequest merges a pull request
func (c *Client) MergePullRequest(id int, sha string, method git.MergeMethod, msg string) error {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d/merge", c.IntegrationConfig.Spec.Git.GetAPIUrl(), url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository), id)

	body := &MergeAcceptRequest{
		Squash:             method == git.MergeMethodSquash,
		Sha:                sha,
		RemoveSourceBranch: false,
	}

	if method == git.MergeMethodSquash {
		body.SquashCommitMessage = msg
	} else {
		body.MergeCommitMessage = msg
	}

	_, _, err := c.requestHTTP(http.MethodPut, apiURL, body)
	if err != nil {
		return err
	}

	return nil
}

// GetPullRequestDiff gets diff of the pull request
func (c *Client) GetPullRequestDiff(id int) (*git.Diff, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d/changes", c.IntegrationConfig.Spec.Git.GetAPIUrl(), url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository), id)

	result, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	rawDiff := &MergeRequestChanges{}
	if err := json.Unmarshal(result, rawDiff); err != nil {
		return nil, err
	}

	var changes []git.Change
	for _, d := range rawDiff.Changes {
		additions, deletions, err := git.GetChangedLinesFromDiff(d.Diff)
		if err != nil {
			return nil, err
		}

		changes = append(changes, git.Change{
			Filename:    d.NewPath,
			OldFilename: d.OldPath,
			Additions:   additions,
			Deletions:   deletions,
			Changes:     additions + deletions,
		})
	}

	return &git.Diff{Changes: changes}, nil
}

// ListPullRequestCommits lists commits list of a pull request
func (c *Client) ListPullRequestCommits(id int) ([]git.Commit, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d/commits", c.IntegrationConfig.Spec.Git.GetAPIUrl(), url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository), id)

	result, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	var resp []CommitResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, err
	}

	var commits []git.Commit
	for _, commit := range resp {
		commits = append(commits, git.Commit{
			SHA:     commit.ID,
			Message: commit.Message,
			Author: git.User{
				Name:  commit.AuthorName,
				Email: commit.AuthorEmail,
			},
			Committer: git.User{
				Name:  commit.CommitterName,
				Email: commit.CommitterEmail,
			},
		})
	}

	return commits, nil
}

// SetLabel sets label to the issue id
func (c *Client) SetLabel(issueType git.IssueType, id int, label string) error {
	var t string
	switch issueType {
	case git.IssueTypeIssue:
		t = "issues"
	case git.IssueTypePullRequest:
		t = "merge_requests"
	default:
		return fmt.Errorf("issue type %s is not supported", issueType)
	}

	apiUrl := fmt.Sprintf("%s/api/v4/projects/%s/%s/%d", c.IntegrationConfig.Spec.Git.GetAPIUrl(), url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository), t, id)

	if _, _, err := c.requestHTTP(http.MethodPut, apiUrl, UpdateMergeRequest{AddLabels: label}); err != nil {
		return err
	}

	return nil
}

// ListLabels lists labels of pr id
func (c *Client) ListLabels(id int) ([]git.IssueLabel, error) {
	apiUrl := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d", c.IntegrationConfig.Spec.Git.GetAPIUrl(), url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository), id)

	raw, _, err := c.requestHTTP(http.MethodGet, apiUrl, nil)
	if err != nil {
		return nil, err
	}

	resp := &MergeRequest{}
	if err := json.Unmarshal(raw, resp); err != nil {
		return nil, err
	}

	var issueLabels []git.IssueLabel
	for _, label := range resp.Labels {
		issueLabels = append(issueLabels, git.IssueLabel{
			Name: label,
		})
	}
	return issueLabels, nil
}

// DeleteLabel deletes label from the issue id
func (c *Client) DeleteLabel(issueType git.IssueType, id int, label string) error {
	var t string
	switch issueType {
	case git.IssueTypeIssue:
		t = "issues"
	case git.IssueTypePullRequest:
		t = "merge_requests"
	default:
		return fmt.Errorf("issue type %s is not supported", issueType)
	}

	apiUrl := fmt.Sprintf("%s/api/v4/projects/%s/%s/%d", c.IntegrationConfig.Spec.Git.GetAPIUrl(), url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository), t, id)

	if _, _, err := c.requestHTTP(http.MethodPut, apiUrl, UpdateMergeRequest{RemoveLabels: label}); err != nil {
		return err
	}
	return nil
}

// GetBranch gets branch info
func (c *Client) GetBranch(branch string) (*git.Branch, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/repository/branches/%s", c.IntegrationConfig.Spec.Git.GetAPIUrl(), url.QueryEscape(c.IntegrationConfig.Spec.Git.Repository), branch)

	raw, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	var resp BranchResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}

	return &git.Branch{Name: resp.Name, CommitID: resp.Commit.ID}, nil
}

func (c *Client) requestHTTP(method, apiURL string, data interface{}) ([]byte, http.Header, error) {
	tlsConfig := c.IntegrationConfig.GetTLSConfig()

	body, header, err := git.RequestHTTP(method, apiURL, c.header, data, tlsConfig)

	if err != nil {
		if isRateLimit, unixTime := CheckRateLimit(string(body), header); isRateLimit {
			rateLimitErr := fmt.Errorf("unixtime::%s. Rate limit exceeded, code %s. Please increase the limit or wait until reset",
				unixTime, strings.Split(strings.Split(err.Error(), ", code ")[1], ",")[0])
			return body, header, rateLimitErr
		}
	}
	return body, header, err
}

// CheckRateLimit checks if the error is a rate limit exceeded error
func CheckRateLimit(msg string, header http.Header) (bool, string) {
	if strings.Contains(msg, "Rate limit exceeded") {
		return true, header.Get("Ratelimit-Reset")
	}
	return false, ""
}

func convertState(original string) git.PullRequestState {
	state := git.PullRequestState(original)
	switch string(state) {
	case "opened":
		state = git.PullRequestStateOpen
	case "closed":
		state = git.PullRequestStateClosed
	}
	return state
}

func convertLabel(original []string) []git.IssueLabel {
	var labels []git.IssueLabel
	for _, l := range original {
		labels = append(labels, git.IssueLabel{Name: l})
	}
	return labels
}

// Validate validates the webhook payload
func Validate(secret, headerToken string) error {
	if secret != headerToken {
		return fmt.Errorf("invalid request : X-Gitlab-Token does not match secret")
	}
	return nil
}
