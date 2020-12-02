package dispatcher

import (
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Dispatcher dispatches IntegrationJob when webhook is called
// A kind of 'plugin' for webhook handler

type Dispatcher struct {
	Client client.Client
}

func (d Dispatcher) Handle(webhook git.Webhook, config *cicdv1.IntegrationConfig) error {
	switch webhook.EventType {
	case git.EventTypePullRequest:
		return d.handlePullRequest(webhook, config)
	case git.EventTypePush:
		return d.handlePush(webhook, config)
	}
	return fmt.Errorf("dispatcher cannot handle event %s", webhook.EventType)
}

func (d Dispatcher) handlePullRequest(webhook git.Webhook, config *cicdv1.IntegrationConfig) error {
	return nil // TODO
}

func (d Dispatcher) handlePush(webhook git.Webhook, config *cicdv1.IntegrationConfig) error {
	return nil // TODO
}
