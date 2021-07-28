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

package trigger

import (
	"context"
	"fmt"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops"
	"github.com/tmax-cloud/cicd-operator/pkg/dispatcher"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Command types for trigger handler
const (
	CommandTypeTest   = "test"
	CommandTypeRetest = "retest"
)

// Handler is an implementation of a ChatOps Handler
type Handler struct {
	Client client.Client
}

// HandleChatOps handles /test and /retest comment commands
func (h *Handler) HandleChatOps(command chatops.Command, webhook *git.Webhook, config *cicdv1.IntegrationConfig) error {
	issueComment := webhook.IssueComment
	// Do nothing if it's not pull request's comment or it's closed
	if issueComment.Issue.PullRequest == nil || issueComment.Issue.PullRequest.State != git.PullRequestStateOpen {
		return nil
	}

	// Authorize or exit
	if err := h.authorize(config, &webhook.Sender, issueComment); err != nil {
		if err := h.registerUnauthorizedComment(config, issueComment.Issue.PullRequest.ID, err); err != nil {
			return err
		}
		return nil
	}

	// Test all (=retest)
	if len(command.Args) == 0 {
		return h.handleRetestCommand(webhook, config)
	}

	return h.handleTestCommand(command, webhook, config)
}

// handleTestCommand handles '/test <ARGS>' command
func (h *Handler) handleTestCommand(command chatops.Command, webhook *git.Webhook, config *cicdv1.IntegrationConfig) error {
	// Generate IntegrationJob for the PullRequest
	prs := []git.PullRequest{*webhook.IssueComment.Issue.PullRequest}
	job := dispatcher.GeneratePreSubmit(prs, &webhook.Repo, &webhook.Sender, config)
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
	if err := h.Client.Create(context.Background(), job); err != nil {
		return err
	}

	return nil
}

// handleTestCommand handles '/retest' command
func (h *Handler) handleRetestCommand(webhook *git.Webhook, config *cicdv1.IntegrationConfig) error {
	// Generate IntegrationJob for the PullRequest
	prs := []git.PullRequest{*webhook.IssueComment.Issue.PullRequest}
	job := dispatcher.GeneratePreSubmit(prs, &webhook.Repo, &webhook.Sender, config)
	if job == nil {
		return nil
	}

	// Create it
	if err := h.Client.Create(context.Background(), job); err != nil {
		return err
	}

	return nil
}

// authorize decides if the sender is authorized to trigger the tests
func (h *Handler) authorize(cfg *cicdv1.IntegrationConfig, sender *git.User, issueComment *git.IssueComment) error {
	// Check if it's PR's author
	if sender.ID == issueComment.Issue.PullRequest.Author.ID {
		return nil
	}

	// Check if it's repo's maintainer
	g, err := utils.GetGitCli(cfg, h.Client)
	if err != nil {
		return err
	}
	ok, err := g.CanUserWriteToRepo(*sender)
	if err != nil {
		return err
	} else if ok {
		return nil
	}

	return &git.UnauthorizedError{User: sender.Name, Repo: cfg.Spec.Git.Repository}
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

// registerUnauthorizedComment registers comment that the user cannot trigger the test
func (h *Handler) registerUnauthorizedComment(config *cicdv1.IntegrationConfig, issueID int, err error) error {
	unAuthErr, ok := err.(*git.UnauthorizedError)
	if !ok {
		return err
	}

	// Skip if token is empty
	if config.Spec.Git.Token == nil {
		return nil
	}

	gitCli, err := utils.GetGitCli(config, h.Client)
	if err != nil {
		return err
	}
	if err := gitCli.RegisterComment(git.IssueTypePullRequest, issueID, generateUnauthorizedComment(unAuthErr.User, unAuthErr.Repo)); err != nil {
		return err
	}
	return nil
}

func generateUnauthorizedComment(user, repo string) string {
	return fmt.Sprintf("User `%s` is not allowed to trigger the test for the repository `%s`\n\n"+
		"If you want to trigger the test, you need to...\n"+
		"- Be author of the pull request\n"+
		"- (For GitHub) Have write permission on the repository\n"+
		"- (For GitLab) Be Developer, Maintainer, or Owner\n", user, repo)
}
