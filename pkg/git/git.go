package git

import (
	"net/http"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
)

type Client interface {
	// Webhooks
	ListWebhook() ([]WebhookEntry, error)
	RegisterWebhook(url string) error
	DeleteWebhook(id int) error
	ParseWebhook(http.Header, []byte) (*Webhook, error)

	// Commit Status
	SetCommitStatus(integrationJob *cicdv1.IntegrationJob, context string, state CommitStatusState, description, targetUrl string) error

	// Users
	GetUserInfo(user string) (*User, error)
}

type CommitStatusState string

type User struct {
	ID    int
	Name  string
	Email string
}
