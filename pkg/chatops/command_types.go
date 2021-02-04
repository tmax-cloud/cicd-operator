package chatops

import (
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
)

type commandType string

const (
	commandTypeTest   = commandType("test")
	commandTypeRetest = commandType("retest")
)

// command is a structure extracted by the comment body
type command struct {
	Type commandType
	Args []string
}

type commandHandler func(command command, webhook *git.Webhook, config *cicdv1.IntegrationConfig) error
