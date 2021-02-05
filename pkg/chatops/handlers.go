package chatops

import (
	"context"
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/dispatcher"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
)

// handleTestCommand handles '/test <ARGS>' command
func (c *ChatOps) handleTestCommand(command command, webhook *git.Webhook, config *cicdv1.IntegrationConfig) error {
	issueComment := webhook.IssueComment
	// Do nothing if it's not pull request's comment or it's closed
	if issueComment.Issue.PullRequest == nil || issueComment.Issue.PullRequest.State != git.PullRequestStateOpen {
		return nil
	}

	// Test all (=retest)
	if len(command.Args) == 0 {
		return c.handleRetestCommand(command, webhook, config)
	}

	if err := c.authorizeUserForTest(webhook); err != nil {
		return err
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
	if err := c.Client.Create(context.Background(), job); err != nil {
		return err
	}

	return nil
}

// handleTestCommand handles '/retest' command
func (c *ChatOps) handleRetestCommand(_ command, webhook *git.Webhook, config *cicdv1.IntegrationConfig) error {
	issueComment := webhook.IssueComment
	// Do nothing if it's not pull request's comment or it's closed
	if issueComment.Issue.PullRequest == nil || issueComment.Issue.PullRequest.State != git.PullRequestStateOpen {
		return nil
	}

	if err := c.authorizeUserForTest(webhook); err != nil {
		return err
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
	if err := c.Client.Create(context.Background(), job); err != nil {
		return err
	}

	return nil
}

// authorizeUserForTest decides if the sender is authorized to trigger the tests
func (c *ChatOps) authorizeUserForTest(webhook *git.Webhook) error {
	issueComment := webhook.IssueComment

	// Check if it's PR's author
	if issueComment.Sender.ID == issueComment.Issue.PullRequest.Sender.ID {
		return nil
	}

	// Check if it's repo's maintainer
	// TODO

	return fmt.Errorf("user %s is not authorized to trigger the test", webhook.IssueComment.Sender.Name)
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
