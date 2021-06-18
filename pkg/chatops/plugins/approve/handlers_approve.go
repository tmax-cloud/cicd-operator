package approve

import (
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// Command types for approve handler
const (
	CommandTypeApprove       = "approve"
	CommandTypeGitLabApprove = "ci-approve"
)

const approvedLabel = "approved"

// Handler is an implementation of both ChatOps Handler and Webhook Plugin for approve
type Handler struct {
	Client client.Client
}

// Handle handles a raw webhook
func (h *Handler) Handle(wh *git.Webhook, ic *cicdv1.IntegrationConfig) error {
	// Skip if token is empty
	if ic.Spec.Git.Token == nil {
		return nil
	}

	// Case 1) Approve / Cancel of a pull request (via github/gitlab feature)
	isApproval := wh.EventType == git.EventTypeIssueComment && wh.IssueComment != nil &&
		wh.IssueComment.Issue.PullRequest.State == git.PullRequestStateOpen && wh.IssueComment.ReviewState != ""

	// Case 2) Label 'approved' is added/deleted
	isLabeled := wh.EventType == git.EventTypePullRequest && wh.PullRequest != nil &&
		(wh.PullRequest.Action == git.PullRequestActionLabeled || wh.PullRequest.Action == git.PullRequestActionUnlabeled)

	// Exit if it's not an approve/cancel action or label action
	if !isApproval && !isLabeled {
		return nil
	}

	gitCli, err := utils.GetGitCli(ic, h.Client)
	if err != nil {
		return err
	}

	// For labeled/unlabeled event
	if isLabeled {
		return h.handleLabelEvent(wh, ic, gitCli)
	}

	// For approve/cancel event
	switch wh.IssueComment.ReviewState {
	case git.PullRequestReviewStateApproved:
		return h.handleApproveCommand(wh.IssueComment, gitCli)
	case git.PullRequestReviewStateUnapproved:
		return h.handleApproveCancelCommand(wh.IssueComment, gitCli)
	}

	return nil
}

// HandleChatOps handles comment commands
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

	// Authorize or exit
	if err := h.authorize(config, webhook.Sender, issueComment.Issue.PullRequest.Author, gitCli); err != nil {
		unAuthErr, ok := err.(*git.UnauthorizedError)
		if !ok {
			return err
		}

		if err := gitCli.RegisterComment(git.IssueTypePullRequest, issueComment.Issue.PullRequest.ID, generateUserUnauthorizedComment(unAuthErr.User)); err != nil {
			return err
		}
		return nil
	}

	// /approve
	if len(command.Args) == 0 {
		return h.handleApproveCommand(issueComment, gitCli)
	}

	// /approve cancel
	if len(command.Args) == 1 && command.Args[0] == "cancel" {
		return h.handleApproveCancelCommand(issueComment, gitCli)
	}

	// Default - malformed comment
	if err := gitCli.RegisterComment(git.IssueTypePullRequest, issueComment.Issue.PullRequest.ID, generateHelpComment()); err != nil {
		return err
	}

	return nil
}

// handleLabelEvent handles labeled/unlabeled event for 'approved' label
func (h *Handler) handleLabelEvent(wh *git.Webhook, ic *cicdv1.IntegrationConfig, gitCli git.Client) error {
	pr := wh.PullRequest
	// Check if 'approved' label is set/unset
	isApprovedChanged := false
	for _, l := range pr.LabelChanged {
		if l.Name == approvedLabel {
			isApprovedChanged = true
			break
		}
	}
	if !isApprovedChanged {
		return nil
	}

	// Is it set or unset?
	// Can't trust pr's action field (gitlab can set/unset labels at the same time)
	isApprovedLabeled := false
	for _, l := range pr.Labels {
		if l.Name == approvedLabel {
			isApprovedLabeled = true
			break
		}
	}

	// Authorize or exit
	if err := h.authorize(ic, wh.Sender, pr.Author, gitCli); err != nil {
		unAuthErr, ok := err.(*git.UnauthorizedError)
		if !ok {
			return err
		}

		// Set/Unset the label again
		if isApprovedLabeled {
			// Delete approved label
			if err := gitCli.DeleteLabel(git.IssueTypePullRequest, pr.ID, approvedLabel); err != nil && !strings.Contains(err.Error(), "Label does not exist") {
				return err
			}
		} else {
			// Register approved label
			if err := gitCli.SetLabel(git.IssueTypePullRequest, pr.ID, approvedLabel); err != nil {
				return err
			}
		}
		if err := gitCli.RegisterComment(git.IssueTypePullRequest, pr.ID, generateUserUnauthorizedComment(unAuthErr.User)); err != nil {
			return err
		}
		return nil
	}
	return nil
}

// handleApproveCommand handles '/approve' command
func (h *Handler) handleApproveCommand(issueComment *git.IssueComment, gitCli git.Client) error {
	// Register approved label
	if err := gitCli.SetLabel(git.IssueTypePullRequest, issueComment.Issue.PullRequest.ID, approvedLabel); err != nil {
		return err
	}

	// Register comment
	if err := gitCli.RegisterComment(git.IssueTypePullRequest, issueComment.Issue.PullRequest.ID, generateApprovedComment(issueComment.Author.Name)); err != nil {
		return err
	}
	return nil
}

// handleApproveCancelCommand handles '/approve cancel] command
func (h *Handler) handleApproveCancelCommand(issueComment *git.IssueComment, gitCli git.Client) error {
	// Delete approved label
	if err := gitCli.DeleteLabel(git.IssueTypePullRequest, issueComment.Issue.PullRequest.ID, approvedLabel); err != nil && !strings.Contains(err.Error(), "Label does not exist") {
		return err
	}

	// Register comment
	if err := gitCli.RegisterComment(git.IssueTypePullRequest, issueComment.Issue.PullRequest.ID, generateApproveCanceledComment(issueComment.Author.Name)); err != nil {
		return err
	}
	return nil
}

// authorize decides if the sender is authorized to approve the PR
func (h *Handler) authorize(cfg *cicdv1.IntegrationConfig, sender git.User, author git.User, gitCli git.Client) error {
	// Check if it's PR's author
	if sender.ID == author.ID {
		return &git.UnauthorizedError{User: sender.Name, Repo: cfg.Spec.Git.Repository}
	}

	// Check if it's repo's maintainer
	ok, err := gitCli.CanUserWriteToRepo(sender)
	if err != nil {
		return err
	} else if ok {
		return nil
	}

	return &git.UnauthorizedError{User: sender.Name, Repo: cfg.Spec.Git.Repository}
}

func generateUserUnauthorizedComment(user string) string {
	return fmt.Sprintf("[APPROVE ALERT]\n\nUser `%s` is not allowed to approve/cancel approve this pull request.\n\n"+
		"Users who meet the following conditions can approve the pull request.\n"+
		"- Not an author of the pull request\n"+
		"- (For GitHub) Have write permission on the repository\n"+
		"- (For GitLab) Be Developer, Maintainer, or Owner\n", user)
}

func generateApprovedComment(user string) string {
	return fmt.Sprintf("[APPROVE ALERT]\n\nUser `%s` approved this pull request!", user)
}

func generateApproveCanceledComment(user string) string {
	return fmt.Sprintf("[APPROVE ALERT]\n\nUser `%s` canceled the approval.", user)
}

func generateHelpComment() string {
	return "[APPROVE ALERT]\n\nApprove comment is malformed\n\n" +
		"You can approve or cancel the approve the pull request by commenting...\n" +
		"- (For GitHub) `/approve`\n" +
		"- (For GitHub) `/approve cancel`\n" +
		"- (For GitLab) `/ci-approve`\n" +
		"- (For GitLab) `/ci-approve cancel`\n"
}
