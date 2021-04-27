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
	SetCommitStatus(sha string, status CommitStatus) error

	// Users

	GetUserInfo(user string) (*User, error)
	CanUserWriteToRepo(user User) (bool, error)

	// Comments

	RegisterComment(issueType IssueType, issueNo int, body string) error

	// Pull Request

	ListPullRequests(onlyOpen bool) ([]PullRequest, error)
	GetPullRequest(id int) (*PullRequest, error)
	MergePullRequest(id int, sha string, method MergeMethod, message string) error

	// Issue Labels

	SetLabel(issueType IssueType, id int, label string) error
	DeleteLabel(issueType IssueType, id int, label string) error

	// Branch

	GetBranch(branch string) (*Branch, error)
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

// MergeMethod is method kind
type MergeMethod string

// MergeMethod types
const (
	MergeMethodSquash = MergeMethod("squash")
	MergeMethodMerge  = MergeMethod("merge")
)
