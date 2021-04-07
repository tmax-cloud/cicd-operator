package chatops

import (
	"context"
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/dispatcher"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
)

// handleTestCommand handles '/test <ARGS>' command
func (c *chatOps) handleTestCommand(command command, webhook *git.Webhook, config *cicdv1.IntegrationConfig) error {
	issueComment := webhook.IssueComment
	// Do nothing if it's not pull request's comment or it's closed
	if issueComment.Issue.PullRequest == nil || issueComment.Issue.PullRequest.State != git.PullRequestStateOpen {
		return nil
	}

	// Test all (=retest)
	if len(command.Args) == 0 {
		return c.handleRetestCommand(command, webhook, config)
	}

	// Authorize or exit
	if err := c.authorizeUserForTest(config, webhook); err != nil {
		if err := c.registerUserUnauthorizedForTestComment(config, issueComment.Issue.PullRequest.ID, err); err != nil {
			return err
		}
		return nil
	}

	// Generate IntegrationJob for the PullRequest
	job, err := dispatcher.GeneratePreSubmit(issueComment.Issue.PullRequest, &webhook.Repo, &issueComment.Sender, config)
	if err != nil {
		return err
	}

	if job == nil {
		return nil
	}

	// Filter only selected (and its dependent) jobs
	if err := filterDependentJobs(command.Args[0], job); err != nil {
		return err
	}

	if len(job.Spec.Jobs) == 0 {
		return nil
	}

	// Create it
	if err := c.client.Create(context.Background(), job); err != nil {
		return err
	}

	return nil
}

// handleTestCommand handles '/retest' command
func (c *chatOps) handleRetestCommand(_ command, webhook *git.Webhook, config *cicdv1.IntegrationConfig) error {
	issueComment := webhook.IssueComment
	// Do nothing if it's not pull request's comment or it's closed
	if issueComment.Issue.PullRequest == nil || issueComment.Issue.PullRequest.State != git.PullRequestStateOpen {
		return nil
	}

	// Authorize or exit
	if err := c.authorizeUserForTest(config, webhook); err != nil {
		if err := c.registerUserUnauthorizedForTestComment(config, issueComment.Issue.PullRequest.ID, err); err != nil {
			return err
		}
		return nil
	}

	// Generate IntegrationJob for the PullRequest
	job, err := dispatcher.GeneratePreSubmit(issueComment.Issue.PullRequest, &webhook.Repo, &issueComment.Sender, config)
	if err != nil {
		return err
	}

	if job == nil {
		return nil
	}

	// Create it
	if err := c.client.Create(context.Background(), job); err != nil {
		return err
	}

	return nil
}

// authorizeUserForTest decides if the sender is authorized to trigger the tests
func (c *chatOps) authorizeUserForTest(cfg *cicdv1.IntegrationConfig, webhook *git.Webhook) error {
	issueComment := webhook.IssueComment

	// Check if it's PR's author
	if issueComment.Sender.ID == issueComment.Issue.PullRequest.Sender.ID {
		return nil
	}

	// Check if it's repo's maintainer
	g, err := utils.GetGitCli(cfg, c.client)
	if err != nil {
		return err
	}
	ok, err := g.CanUserWriteToRepo(issueComment.Sender)
	if err != nil {
		return err
	} else if ok {
		return nil
	}

	return &unauthorizedError{user: issueComment.Sender.Name, repo: cfg.Spec.Git.Repository}
}

// filterDependentJobs filters out unnecessary (not dependent) jobs
func filterDependentJobs(target string, job *cicdv1.IntegrationJob) error {
	dependents, err := dependentJobs(target, job.Spec.Jobs)
	if err != nil {
		return err
	}

	filteredJobs := cicdv1.Jobs{}
	for _, j := range job.Spec.Jobs {
		if _, depends := dependents[j.Name]; depends {
			filteredJobs = append(filteredJobs, j)
		}
	}

	job.Spec.Jobs = filteredJobs

	return nil
}

// dependentJobs find all the dependent jobs for the job
// It looks for job.After field
func dependentJobs(wanted string, jobs cicdv1.Jobs) (map[string]struct{}, error) {
	depends := map[string]struct{}{}
	depends[wanted] = struct{}{}

	graph, err := jobs.GetGraph()
	if err != nil {
		return nil, err
	}

	pres := graph.GetPres(wanted)
	for _, p := range pres {
		depends[p] = struct{}{}
	}
	return depends, nil
}

// registerUserUnauthorizedForTestComment registers comment that the user cannot trigger the test
func (c *chatOps) registerUserUnauthorizedForTestComment(config *cicdv1.IntegrationConfig, issueID int, err error) error {
	unAuthErr, ok := err.(*unauthorizedError)
	if !ok {
		return err
	}

	// Skip if token is empty
	if config.Spec.Git.Token == nil {
		return nil
	}

	gitCli, err := utils.GetGitCli(config, c.client)
	if err != nil {
		return err
	}
	if err := gitCli.RegisterComment(git.IssueTypePullRequest, issueID, generateUserUnauthorizedForTestComment(unAuthErr.user, unAuthErr.repo)); err != nil {
		return err
	}
	return nil
}

func generateUserUnauthorizedForTestComment(user, repo string) string {
	return fmt.Sprintf("User `%s` is not allowed to trigger the test for the repository `%s`\n\n"+
		"If you want to trigger the test, you need to...\n"+
		"- Be author of the pull request\n"+
		"- (For GitHub) Have write permission on the repository\n"+
		"- (For GitLab) Be Developer, Maintainer, or Owner\n", user, repo)
}
