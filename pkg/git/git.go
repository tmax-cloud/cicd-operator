package git

import (
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"net/http"
)

type Client interface {
	// Webhooks
	RegisterWebhook(gitConfig *cicdv1.GitConfig, url string) error
	ParseWebhook(http.Header, []byte) (Webhook, error)

	// Commit Status
	SetCommitStatus(gitConfig *cicdv1.GitConfig, context string, state CommitStatusState, description, targetUrl string) error
}

type CommitStatusState string

const (
	CommitStatusStateSuccess = CommitStatusState("success")
	CommitStatusStateFailure = CommitStatusState("failure")
	CommitStatusStateError   = CommitStatusState("error")
	CommitStatusStatePending = CommitStatusState("pending")
)
