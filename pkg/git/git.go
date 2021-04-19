package git

import (
	"net/http"
)

// Client is a git client interface
type Client interface {
	Init() error

	// Webhooks

	ListWebhook() ([]WebhookEntry, error)
	RegisterWebhook(url string) error
	DeleteWebhook(id int) error
	ParseWebhook(http.Header, []byte) (*Webhook, error)

	// Commit Status

	ListCommitStatuses(ref string) ([]CommitStatus, error)
	SetCommitStatus(sha, context string, state CommitStatusState, description, targetURL string) error

	// Users

	GetUserInfo(user string) (*User, error)
	CanUserWriteToRepo(user User) (bool, error)

	// Comments

	RegisterComment(issueType IssueType, issueNo int, body string) error

	// Pull Request

	ListPullRequests(onlyOpen bool) ([]PullRequest, error)
	GetPullRequest(id int) (*PullRequest, error)
}

// IssueType is a type of the issue
type IssueType string

// IssueType constants
const (
	IssueTypeIssue       = IssueType("issue")
	IssueTypePullRequest = IssueType("pull_request")
)

// CommitStatusState is a commit status type
type CommitStatusState string

// CommitStatusStates
const (
	CommitStatusStateSuccess = CommitStatusState("success")
	CommitStatusStateFailure = CommitStatusState("failure")
	CommitStatusStateError   = CommitStatusState("error")
	CommitStatusStatePending = CommitStatusState("pending")
)

// FakeSha is a fake SHA for a commit
const (
	FakeSha = "0000000000000000000000000000000000000000"
)
