package git

import (
	"net/http"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	// Webhooks
	ListWebhook(integrationConfig *cicdv1.IntegrationConfig, c client.Client) ([]WebhookEntry, error)
	RegisterWebhook(integrationConfig *cicdv1.IntegrationConfig, url string, c client.Client) error
	DeleteWebhook(integrationConfig *cicdv1.IntegrationConfig, id int, c client.Client) error
	ParseWebhook(*cicdv1.IntegrationConfig, http.Header, []byte) (Webhook, error)

	// Commit Status
	SetCommitStatus(integrationJob *cicdv1.IntegrationJob, integrationConfig *cicdv1.IntegrationConfig, context string, state CommitStatusState, description, targetUrl string, c client.Client) error
}

type CommitStatusState string
