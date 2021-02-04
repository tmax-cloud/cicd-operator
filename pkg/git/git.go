package git

import (
	"net/http"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
)

// Client is a git client interface
type Client interface {
	// Webhooks
	ListWebhook() ([]WebhookEntry, error)
	RegisterWebhook(url string) error
	DeleteWebhook(id int) error
	ParseWebhook(http.Header, []byte) (*Webhook, error)

	// Commit Status
	SetCommitStatus(integrationJob *cicdv1.IntegrationJob, context string, state CommitStatusState, description, targetURL string) error

	// Users
	GetUserInfo(user string) (*User, error)
}

// CommitStatusState is a commit status type
type CommitStatusState string
