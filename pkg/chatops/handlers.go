package chatops

import (
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
)

// handleTestCommand handles '/test <ARGS>' command
func (c *ChatOps) handleTestCommand(command command, webhook *git.Webhook, config *cicdv1.IntegrationConfig) error {
	// Do nothing if it's not pull request's comment
	if webhook.IssueComment.Issue.PullRequest == nil {
		return nil
	}

	// Test all (=retest)
	if len(command.Args) == 0 {
		return c.handleRetestCommand(command, webhook, config)
	}

	if err := c.authorizeUserForTest(webhook); err != nil {
		return err
	}

	// TODO

	return nil
}

// handleTestCommand handles '/retest' command
func (c *ChatOps) handleRetestCommand(command command, webhook *git.Webhook, config *cicdv1.IntegrationConfig) error {
	// Do nothing if it's not pull request's comment
	if webhook.IssueComment.Issue.PullRequest == nil {
		return nil
	}

	if err := c.authorizeUserForTest(webhook); err != nil {
		return err
	}

	// TODO

	return nil
}

// authorizeUserForTest decides if the sender is authorized to trigger the tests
func (c *ChatOps) authorizeUserForTest(webhook *git.Webhook) error {
	// TODO - create IntegrationJob

	// return fmt.Errorf("user %s is not authorized to trigger a test", webhook.IssueComment.Sender.Name)
	return nil
}
