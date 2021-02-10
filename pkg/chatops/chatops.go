package chatops

import (
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// chatOps triggers tests/retests via comments
type chatOps struct {
	client   client.Client
	handlers map[commandType]commandHandler
}

// New is a constructor fo chatOps
func New(c client.Client) *chatOps {
	co := &chatOps{
		client:   c,
		handlers: map[commandType]commandHandler{},
	}

	co.registerCommandHandler(commandTypeTest, co.handleTestCommand)
	co.registerCommandHandler(commandTypeRetest, co.handleRetestCommand)

	return co
}

// Handle actually handles the webhook payload to create IntegrationJob
func (c *chatOps) Handle(webhook *git.Webhook, config *cicdv1.IntegrationConfig) error {
	issueComment := webhook.IssueComment
	if issueComment == nil {
		return nil
	}

	// Extract commands from comment and call handler
	commands := c.extractCommands(issueComment.Comment.Body)
	for _, command := range commands {
		handler, ok := c.handlers[command.Type]
		if !ok {
			continue
		}
		if err := handler(command, webhook, config); err != nil {
			return err
		}
	}

	return nil
}

// extractCommands extracts commands (i.e. /[a-z], e.g., /test /retest /assign) from the comment body
func (c *chatOps) extractCommands(comment string) []command {
	var commands []command

	lines := strings.Split(comment, "\n")

	for _, l := range lines {
		if len(l) > 2 && l[0] == '/' && 'a' <= l[1] && l[1] <= 'z' {
			tokens := strings.Split(l, " ")
			commands = append(commands, command{
				Type: commandType(tokens[0][1:]),
				Args: tokens[1:],
			})
		}
	}

	return commands
}

func (c *chatOps) registerCommandHandler(command commandType, handler commandHandler) {
	c.handlers[command] = handler
}
