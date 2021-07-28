/*
 Copyright 2021 The CI/CD Operator Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

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
	handlers map[string]CommandHandler
}

// New is a constructor fo chatOps
func New(c client.Client) *chatOps {
	co := &chatOps{
		client:   c,
		handlers: map[string]CommandHandler{},
	}

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
func (c *chatOps) extractCommands(comment string) []Command {
	var commands []Command

	lines := strings.Split(comment, "\n")

	for _, l := range lines {
		if len(l) > 2 && l[0] == '/' && 'a' <= l[1] && l[1] <= 'z' {
			tokens := strings.Split(l, " ")
			commands = append(commands, Command{
				Type: tokens[0][1:],
				Args: tokens[1:],
			})
		}
	}

	return commands
}

func (c *chatOps) RegisterCommandHandler(command string, handler CommandHandler) {
	c.handlers[command] = handler
}
