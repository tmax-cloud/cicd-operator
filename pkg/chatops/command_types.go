package chatops

import (
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
)

// Command is a structure extracted by the comment body
type Command struct {
	Type string
	Args []string
}

// CommandHandler is a handler function type for chat ops events
type CommandHandler func(command Command, webhook *git.Webhook, config *cicdv1.IntegrationConfig) error
