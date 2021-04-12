package git

import (
	"net/http"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
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
	SetCommitStatus(integrationJob *cicdv1.IntegrationJob, context string, state CommitStatusState, description, targetURL string) error

	// Users

	GetUserInfo(user string) (*User, error)
	CanUserWriteToRepo(user User) (bool, error)

	// Comments

	RegisterComment(issueType IssueType, issueNo int, body string) error

	// Pull Request

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

// FakeSha is a fake SHA for a commit
const (
	FakeSha = "0000000000000000000000000000000000000000"
)
