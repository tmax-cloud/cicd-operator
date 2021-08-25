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

	var entries []WebhookEntry
	err := git.GetPaginatedRequest(apiURL, c.header, func() interface{} {
		return &[]WebhookEntry{}
	}, func(i interface{}) {
		entries = append(entries, *i.(*[]WebhookEntry)...)
	})
	if err != nil {
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

	var statuses []CommitStatusResponse
	err := git.GetPaginatedRequest(apiURL, c.header, func() interface{} {
		return &[]CommitStatusResponse{}
	}, func(i interface{}) {
		statuses = append(statuses, *i.(*[]CommitStatusResponse)...)
	})
	if err != nil {
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
			Context:     s.Context,
			State:       git.CommitStatusState(s.State),
			Description: s.Description,
			TargetURL:   s.TargetURL,
		})
	}

	return resp, nil
}

// SetCommitStatus sets commit status for the specific commit
func (c *Client) SetCommitStatus(sha string, status git.CommitStatus) error {
	var commitStatusBody CommitStatusRequest

	// Don't set commit status if its' sha is a fake
	if sha == git.FakeSha {
		return nil
	}

	apiURL := c.IntegrationConfig.Spec.Git.GetAPIUrl() + "/repos/" + c.IntegrationConfig.Spec.Git.Repository + "/statuses/" + sha

	commitStatusBody.State = string(status.State)
	commitStatusBody.TargetURL = status.TargetURL
	commitStatusBody.Description = status.Description
	commitStatusBody.Context = status.Context

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

	var prs []PullRequest
	err := git.GetPaginatedRequest(apiURL, c.header, func() interface{} {
		return &[]PullRequest{}
	}, func(i interface{}) {
		prs = append(prs, *i.(*[]PullRequest)...)
	})
	if err != nil {
		return nil, err
	}

	var result []git.PullRequest
	for _, pr := range prs {
		if !pr.Draft { // TODO - should it be here??
			result = append(result, *convertPullRequestToShared(&pr))
		}
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

// MergePullRequest merges a pull request
func (c *Client) MergePullRequest(id int, sha string, method git.MergeMethod, message string) error {
	apiURL := fmt.Sprintf("%s/repos/%s/pulls/%d/merge", c.IntegrationConfig.Spec.Git.GetAPIUrl(), c.IntegrationConfig.Spec.Git.Repository, id)

	tokens := strings.Split(message, "\n\n")

	body := &MergeRequest{
		CommitTitle: tokens[0],
		MergeMethod: string(method),
		Sha:         sha,
	}

	if len(tokens) > 1 {
		body.CommitMessage = strings.Join(tokens[1:], "\n\n")
	}

	_, _, err := c.requestHTTP(http.MethodPut, apiURL, body)
	if err != nil {
		return err
	}

	return nil
}

// GetPullRequestDiff gets diff of the pull request
func (c *Client) GetPullRequestDiff(id int) (*git.Diff, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/pulls/%d/files", c.IntegrationConfig.Spec.Git.GetAPIUrl(), c.IntegrationConfig.Spec.Git.Repository, id)
	rawDiffs, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	diffs := DiffFiles{}
	if err := json.Unmarshal(rawDiffs, &diffs); err != nil {
		return nil, err
	}

	var changes []git.Change
	for _, d := range diffs {
		prevName := d.PrevFilename
		if prevName == "" {
			prevName = d.Filename
		}
		changes = append(changes, git.Change{
			Filename:    d.Filename,
			OldFilename: prevName,
			Additions:   d.Additions,
			Deletions:   d.Deletions,
			Changes:     d.Changes,
		})
	}

	return &git.Diff{Changes: changes}, nil
}

// ListPullRequestCommits lists commits list of a pull request
func (c *Client) ListPullRequestCommits(id int) ([]git.Commit, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/pulls/%d/commits", c.IntegrationConfig.Spec.Git.GetAPIUrl(), c.IntegrationConfig.Spec.Git.Repository, id)

	raw, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	var resp []CommitResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}

	var commits []git.Commit
	for _, commit := range resp {
		commits = append(commits, git.Commit{
			SHA:     commit.SHA,
			Message: commit.Commit.Message,
			Author: git.User{
				Name:  commit.Commit.Author.Name,
				Email: commit.Commit.Author.Email,
			},
			Committer: git.User{
				Name:  commit.Commit.Committer.Name,
				Email: commit.Commit.Committer.Email,
			},
		})
	}

	return commits, nil
}

// SetLabel sets label to the issue id
func (c *Client) SetLabel(_ git.IssueType, id int, label string) error {
	apiURL := fmt.Sprintf("%s/repos/%s/issues/%d/labels", c.IntegrationConfig.Spec.Git.GetAPIUrl(), c.IntegrationConfig.Spec.Git.Repository, id)

	_, _, err := c.requestHTTP(http.MethodPost, apiURL, []LabelBody{{Name: label}})
	if err != nil {
		return err
	}

	return nil
}

// DeleteLabel deletes label from the issue id
func (c *Client) DeleteLabel(_ git.IssueType, id int, label string) error {
	apiURL := fmt.Sprintf("%s/repos/%s/issues/%d/labels/%s", c.IntegrationConfig.Spec.Git.GetAPIUrl(), c.IntegrationConfig.Spec.Git.Repository, id, label)

	_, _, err := c.requestHTTP(http.MethodDelete, apiURL, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetBranch gets branch info
func (c *Client) GetBranch(branch string) (*git.Branch, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/branches/%s", c.IntegrationConfig.Spec.Git.GetAPIUrl(), c.IntegrationConfig.Spec.Git.Repository, branch)

	raw, _, err := c.requestHTTP(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp := &BranchResponse{}
	if err := json.Unmarshal(raw, resp); err != nil {
		return nil, err
	}

	return &git.Branch{Name: resp.Name, CommitID: resp.Commit.Sha}, nil
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
		Author: git.User{
			ID:   pr.User.ID,
			Name: pr.User.Name,
		},
		URL:       pr.URL,
		Base:      git.Base{Ref: pr.Base.Ref, Sha: pr.Base.Sha},
		Head:      git.Head{Ref: pr.Head.Ref, Sha: pr.Head.Sha},
		Labels:    labels,
		Mergeable: pr.Mergeable,
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
