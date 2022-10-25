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

package hold

import (
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

// CommandTypeHold is a hold command type
const (
	CommandTypeHold = "hold"
)

var log = logf.Log.WithName("hold-plugin")

// Handler is an implementation of a ChatOps Handler
type Handler struct {
	Client client.Client
}

// HandleChatOps handles /hold and /hold cancel comment commands
func (h *Handler) HandleChatOps(command chatops.Command, webhook *git.Webhook, config *cicdv1.IntegrationConfig) error {
	issueComment := webhook.IssueComment
	// Do nothing if it's not pull request's comment or it's closed
	if issueComment.Issue.PullRequest == nil || issueComment.Issue.PullRequest.State != git.PullRequestStateOpen {
		return nil
	}

	// Skip if token is empty
	if config.Spec.Git.Token == nil {
		return nil
	}

	gitCli, err := utils.GetGitCli(config, h.Client)
	if err != nil {
		return err
	}

	// /hold
	if len(command.Args) == 0 {
		return h.handleHoldCommand(issueComment, gitCli)
	}

	// /hold cancel
	if len(command.Args) == 1 && command.Args[0] == "cancel" {
		return h.handleHoldCancelCommand(issueComment, gitCli)
	}

	// Default - malformed comment
	if err := gitCli.RegisterComment(git.IssueTypePullRequest, issueComment.Issue.PullRequest.ID, "", generateHelpComment()); err != nil {
		return err
	}

	return nil
}

// handleHoldCommand handles '/hold' command
func (h *Handler) handleHoldCommand(issueComment *git.IssueComment, gitCli git.Client) error {
	log.Info(fmt.Sprintf("%s held %s", issueComment.Author.Name, issueComment.Issue.PullRequest.URL))
	// Register hold label
	if err := gitCli.SetLabel(git.IssueTypePullRequest, issueComment.Issue.PullRequest.ID, configs.MergeBlockLabel); err != nil {
		return err
	}
	return nil
}

// handleHoldCancelCommand handles '/hold cancel' command
func (h *Handler) handleHoldCancelCommand(issueComment *git.IssueComment, gitCli git.Client) error {
	log.Info(fmt.Sprintf("%s canceled hold on %s", issueComment.Author.Name, issueComment.Issue.PullRequest.URL))
	// Delete hold label
	if err := gitCli.DeleteLabel(git.IssueTypePullRequest, issueComment.Issue.PullRequest.ID, configs.MergeBlockLabel); err != nil && !strings.Contains(err.Error(), "Label does not exist") {
		return err
	}
	return nil
}

func generateHelpComment() string {
	return "[HOLD ALERT]\n\nHold comment is malformed\n\n" +
		"You can hold or cancel hold the pull request by commenting...\n" +
		"- `/hold`\n" +
		"- `/hold cancel`\n"
}
