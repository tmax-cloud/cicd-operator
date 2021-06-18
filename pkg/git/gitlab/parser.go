package gitlab

import (
	"encoding/json"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)

func (c *Client) parsePullRequestWebhook(jsonString []byte) (*git.Webhook, error) {
	var data MergeRequestWebhook
	if err := json.Unmarshal(jsonString, &data); err != nil {
		return nil, err
	}

	// Sender - Author might be different... in case of update
	sender := git.User{Name: data.User.Name, Email: data.User.Email}
	var author git.User
	if data.User.ID == data.ObjectAttribute.AuthorID {
		author = sender
	} else {
		user, err := c.GetUserInfo(strconv.Itoa(data.ObjectAttribute.AuthorID))
		if err != nil {
			return nil, err
		}
		author = *user
	}

	base := git.Base{Ref: data.ObjectAttribute.BaseRef}
	head := git.Head{Ref: data.ObjectAttribute.HeadRef, Sha: data.ObjectAttribute.LastCommit.Sha}
	repo := git.Repository{Name: data.Project.Name, URL: data.Project.WebURL}
	action := git.PullRequestAction(data.ObjectAttribute.Action)
	switch string(action) {
	case "close":
		action = git.PullRequestActionClose
	case "open":
		action = git.PullRequestActionOpen
	case "reopen":
		action = git.PullRequestActionReOpen
	case "update":
		if data.ObjectAttribute.OldRev != "" {
			action = git.PullRequestActionSynchronize
		} else if data.Changes.Labels != nil {
			action = git.PullRequestActionLabeled // Maybe unlabeled... but doesn't matter
		}
	case "approved", "unapproved":
		return c.parsePullRequestReviewWebhook(data)
	}

	// Get Target branch
	baseBranch, err := c.GetBranch(data.ObjectAttribute.BaseRef)
	if err != nil {
		return nil, err
	}
	base.Sha = baseBranch.CommitID

	state := git.PullRequestState(data.ObjectAttribute.State)
	switch string(state) {
	case "opened":
		state = git.PullRequestStateOpen
	case "closed":
		state = git.PullRequestStateClosed
	}

	var labels []git.IssueLabel
	for _, l := range data.Labels {
		labels = append(labels, git.IssueLabel{Name: l.Title})
	}

	pullRequest := git.PullRequest{ID: data.ObjectAttribute.ID, Title: data.ObjectAttribute.Title, Author: author, URL: data.Project.WebURL, Base: base, Head: head, State: state, Action: action, Labels: labels}
	return &git.Webhook{EventType: git.EventTypePullRequest, Repo: repo, Sender: sender, PullRequest: &pullRequest}, nil
}

func (c *Client) parsePushWebhook(jsonString []byte) (*git.Webhook, error) {
	var data PushWebhook

	if err := json.Unmarshal(jsonString, &data); err != nil {
		return nil, err
	}
	repo := git.Repository{Name: data.Project.Name, URL: data.Project.WebURL}
	if strings.HasPrefix(data.Sha, "0000") && strings.HasSuffix(data.Sha, "0000") {
		return nil, nil
	}
	sender := git.User{Name: data.UserName, ID: data.UserID}
	push := git.Push{Ref: data.Ref, Sha: data.Sha}

	// Get sender email
	userInfo, err := c.GetUserInfo(strconv.Itoa(data.UserID))
	if err == nil {
		sender.Email = userInfo.Email
	}

	return &git.Webhook{EventType: git.EventTypePush, Repo: repo, Sender: sender, Push: &push}, nil
}

func (c *Client) parseIssueComment(jsonString []byte) (*git.Webhook, error) {
	data := &NoteHook{}

	if err := json.Unmarshal(jsonString, data); err != nil {
		return nil, err
	}

	mrState := git.PullRequestState(data.MergeRequest.State)
	switch string(mrState) {
	case "opened":
		mrState = git.PullRequestStateOpen
	case "closed":
		mrState = git.PullRequestStateClosed
	}

	// Only handle creation
	if !data.ObjectAttributes.CreatedAt.Time.Equal(data.ObjectAttributes.UpdatedAt.Time) {
		return nil, nil
	}

	sender := git.User{
		ID:    data.User.ID,
		Name:  data.User.Name,
		Email: data.User.Email,
	}
	var author git.User
	if sender.ID == data.ObjectAttributes.AuthorID {
		author = sender
	} else {
		user, err := c.GetUserInfo(strconv.Itoa(data.ObjectAttributes.AuthorID))
		if err != nil {
			return nil, err
		}
		author = *user
	}

	// Get Merge Request user info
	var pr *git.PullRequest
	if data.MergeRequest.TargetBranch != "" {
		// Get User info
		mrAuthor, err := c.GetUserInfo(strconv.Itoa(data.MergeRequest.AuthorID))
		if err != nil {
			mrAuthor = &git.User{ID: data.MergeRequest.AuthorID}
		}
		// Get Target branch
		baseBranch, err := c.GetBranch(data.MergeRequest.TargetBranch)
		if err != nil {
			return nil, err
		}
		pr = &git.PullRequest{
			ID:     data.MergeRequest.ID,
			Title:  data.MergeRequest.Title,
			State:  mrState,
			Author: *mrAuthor,
			URL:    data.MergeRequest.URL,
			Base: git.Base{
				Ref: data.MergeRequest.TargetBranch,
				Sha: baseBranch.CommitID,
			},
			Head: git.Head{
				Ref: data.MergeRequest.SourceBranch,
				Sha: data.MergeRequest.LastCommit.ID,
			},
		}
	}

	return &git.Webhook{EventType: git.EventTypeIssueComment, Repo: git.Repository{
		Name: data.Project.Name,
		URL:  data.Project.WebURL,
	},
		Sender: sender,
		IssueComment: &git.IssueComment{
			Comment: git.Comment{
				Body:      data.ObjectAttributes.Note,
				CreatedAt: &metav1.Time{Time: data.ObjectAttributes.CreatedAt.Time},
			},
			Issue: git.Issue{
				PullRequest: pr,
			},
			Author: author,
		}}, nil
}

func (c *Client) parsePullRequestReviewWebhook(data MergeRequestWebhook) (*git.Webhook, error) {
	state := git.PullRequestState(data.ObjectAttribute.State)
	switch string(state) {
	case "opened":
		state = git.PullRequestStateOpen
	case "closed":
		state = git.PullRequestStateClosed
	}

	sender := git.User{
		ID:    data.User.ID,
		Name:  data.User.Name,
		Email: data.User.Email,
	}
	commentAuthor := sender

	// Get User info
	mrAuthor, err := c.GetUserInfo(strconv.Itoa(data.ObjectAttribute.AuthorID))
	if err != nil {
		mrAuthor = &git.User{ID: data.ObjectAttribute.AuthorID}
	}
	// Get Target branch
	baseBranch, err := c.GetBranch(data.ObjectAttribute.BaseRef)
	if err != nil {
		return nil, err
	}

	reviewState := git.PullRequestReviewState(data.ObjectAttribute.Action)
	switch reviewState {
	case "approved":
		reviewState = git.PullRequestReviewStateApproved
	case "unapproved":
		reviewState = git.PullRequestReviewStateUnapproved
	}

	return &git.Webhook{
		EventType: git.EventTypePullRequestReview,
		Repo: git.Repository{
			Name: data.Project.Name,
			URL:  data.Project.WebURL,
		},
		Sender: sender,
		IssueComment: &git.IssueComment{
			Author:      commentAuthor,
			ReviewState: reviewState,
			Issue: git.Issue{
				PullRequest: &git.PullRequest{
					ID:     data.ObjectAttribute.ID,
					Title:  data.ObjectAttribute.Title,
					Author: *mrAuthor,
					URL:    data.Project.WebURL,
					Base:   git.Base{Ref: data.ObjectAttribute.BaseRef, Sha: baseBranch.CommitID},
					Head:   git.Head{Ref: data.ObjectAttribute.HeadRef, Sha: data.ObjectAttribute.LastCommit.Sha},
					State:  state,
				},
			},
		},
	}, nil
}
