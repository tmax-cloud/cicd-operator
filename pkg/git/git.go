package git

import (
	"net/http"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	// Webhooks
	RegisterWebhook(integrationConfig *cicdv1.IntegrationConfig, url string, c *client.Client) error
	ParseWebhook(http.Header, []byte) (Webhook, error)

	// Commit Status
	SetCommitStatus(gitConfig *cicdv1.GitConfig, context string, state CommitStatusState, token, description, targetUrl string) error
}

type CommitStatusState string

const (
	CommitStatusStateSuccess = CommitStatusState("success")
	CommitStatusStateFailure = CommitStatusState("failure")
	CommitStatusStateError   = CommitStatusState("error")
	CommitStatusStatePending = CommitStatusState("pending")
)
