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

	sender, author, err := c.getSenderAuthor(data.User, data.ObjectAttribute.AuthorID)
	if err != nil {
		return nil, err
	}

	pullRequest := git.PullRequest{ID: data.ObjectAttribute.ID, Title: data.ObjectAttribute.Title, URL: data.Project.WebURL}
	pullRequest.Author = *author
	pullRequest.Base = git.Base{Ref: data.ObjectAttribute.BaseRef}
	pullRequest.Head = git.Head{Ref: data.ObjectAttribute.HeadRef, Sha: data.ObjectAttribute.LastCommit.Sha}
	repo := git.Repository{Name: data.Project.Name, URL: data.Project.WebURL}
	pullRequest.Action = git.PullRequestAction(data.ObjectAttribute.Action)
	switch string(pullRequest.Action) {
	case "close":
		pullRequest.Action = git.PullRequestActionClose
	case "open":
		pullRequest.Action = git.PullRequestActionOpen
	case "reopen":
		pullRequest.Action = git.PullRequestActionReOpen
	case "update":
		if data.ObjectAttribute.OldRev != "" {
			pullRequest.Action = git.PullRequestActionSynchronize
		} else if data.Changes.Labels != nil {
			var isUnlabeled bool
			pullRequest.LabelChanged, isUnlabeled = diffLabels(data.Changes.Labels.Previous, data.Changes.Labels.Current)
			pullRequest.Action = git.PullRequestActionLabeled
			if isUnlabeled {
				pullRequest.Action = git.PullRequestActionUnlabeled
			}
		}
	case "approved", "unapproved":
		return c.parsePullRequestReviewWebhook(data)
	}

	// Get Target branch
	baseBranch, err := c.GetBranch(data.ObjectAttribute.BaseRef)
	if err != nil {
		return nil, err
	}
	pullRequest.Base.Sha = baseBranch.CommitID

	pullRequest.State = git.PullRequestState(data.ObjectAttribute.State)
	switch string(pullRequest.State) {
	case "opened":
		pullRequest.State = git.PullRequestStateOpen
	case "closed":
		pullRequest.State = git.PullRequestStateClosed
	}

	for _, l := range data.Labels {
		pullRequest.Labels = append(pullRequest.Labels, git.IssueLabel{Name: l.Title})
	}

	return &git.Webhook{EventType: git.EventTypePullRequest, Repo: repo, PullRequest: &pullRequest, Sender: *sender}, nil
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

	var mrState git.PullRequestState
	var pr *git.PullRequest

	mr := NoteMR{}
	if data.MergeRequest != mr {
		mrState = git.PullRequestState(data.MergeRequest.State)
		switch string(mrState) {
		case "opened":
			mrState = git.PullRequestStateOpen
		case "closed":
			mrState = git.PullRequestStateClosed
		}
		// Get Merge Request user info
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
	}

	// Only handle creation
	if !data.ObjectAttributes.CreatedAt.Time.Equal(data.ObjectAttributes.UpdatedAt.Time) {
		return nil, nil
	}

	sender, author, err := c.getSenderAuthor(data.User, data.ObjectAttributes.AuthorID)
	if err != nil {
		return nil, err
	}

	return &git.Webhook{EventType: git.EventTypeIssueComment, Repo: git.Repository{
		Name: data.Project.Name,
		URL:  data.Project.WebURL,
	},
		Sender: *sender,
		IssueComment: &git.IssueComment{
			Comment: git.Comment{
				Body:      data.ObjectAttributes.Note,
				CreatedAt: &metav1.Time{Time: data.ObjectAttributes.CreatedAt.Time},
			},
			Issue: git.Issue{
				PullRequest: pr,
				CommitID:    data.Commit.ID,
			},
			Author: *author,
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

func (c *Client) getSenderAuthor(senderPre User, authorID int) (*git.User, *git.User, error) {
	sender := git.User{ID: senderPre.ID, Name: senderPre.Name, Email: senderPre.Email}
	var author git.User
	if sender.ID == authorID {
		author = sender
	} else {
		user, err := c.GetUserInfo(strconv.Itoa(authorID))
		if err != nil {
			return nil, nil, err
		}
		author = *user
	}

	return &sender, &author, nil
}

func diffLabels(prev, cur []Label) ([]git.IssueLabel, bool) {
	var diff []git.IssueLabel
	isUnlabeled := false

	prevMap := map[string]Label{}
	curMap := map[string]Label{}

	for _, l := range prev {
		prevMap[l.Title] = l
	}
	for _, l := range cur {
		curMap[l.Title] = l
	}

	for _, l := range prev {
		if _, exist := curMap[l.Title]; !exist {
			diff = append(diff, git.IssueLabel{Name: l.Title})
		}
	}
	for _, l := range cur {
		if _, exist := prevMap[l.Title]; !exist {
			diff = append(diff, git.IssueLabel{Name: l.Title})
			isUnlabeled = true
		}
	}

	return diff, isUnlabeled
}
