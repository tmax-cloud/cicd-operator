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

package github

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/tmax-cloud/cicd-operator/pkg/git"
)

func (c *Client) parsePullRequestWebhook(jsonString []byte) (*git.Webhook, error) {
	var data PullRequestWebhook
	if err := json.Unmarshal(jsonString, &data); err != nil {
		return nil, err
	}

	pullRequest := git.PullRequest{ID: data.Number, Title: data.PullRequest.Title, URL: data.Repo.URL, State: git.PullRequestState(data.PullRequest.State), Action: git.PullRequestAction(data.Action)}

	// Get sender & author
	sender, author := c.getSenderAuthor(data.Sender, data.PullRequest.User)
	pullRequest.Author = *author

	for _, l := range data.PullRequest.Labels {
		pullRequest.Labels = append(pullRequest.Labels, git.IssueLabel{Name: l.Name})
	}

	// Labeled/Unlabeled event
	if data.Action == string(git.PullRequestActionLabeled) || data.Action == string(git.PullRequestActionUnlabeled) {
		pullRequest.LabelChanged = append(pullRequest.LabelChanged, git.IssueLabel{Name: data.Label.Name})
	}

	pullRequest.Base = git.Base{Ref: data.PullRequest.Base.Ref, Sha: data.PullRequest.Base.Sha}
	pullRequest.Head = git.Head{Ref: data.PullRequest.Head.Ref, Sha: data.PullRequest.Head.Sha}
	repo := git.Repository{Name: data.Repo.Name, URL: data.Repo.URL}
	return &git.Webhook{EventType: git.EventTypePullRequest, Repo: repo, PullRequest: &pullRequest, Sender: *sender, RequestBody: string(jsonString)}, nil
}

func (c *Client) parsePushWebhook(jsonString []byte) (*git.Webhook, error) {
	var data PushWebhook

	if err := json.Unmarshal(jsonString, &data); err != nil {
		return nil, err
	}
	repo := git.Repository{Name: data.Repo.Name, URL: data.Repo.URL}
	if strings.HasPrefix(data.Sha, "0000") && strings.HasSuffix(data.Sha, "0000") {
		return nil, nil
	}
	sender := git.User{Name: data.Sender.Name, ID: data.Sender.ID}
	push := git.Push{Ref: data.Ref, Sha: data.Sha}

	// Get sender email
	userInfo, err := c.GetUserInfo(data.Sender.Name)
	if err == nil {
		sender.Email = userInfo.Email
	}

	return &git.Webhook{EventType: git.EventTypePush, Repo: repo, Sender: sender, Push: &push, RequestBody: string(jsonString)}, nil
}

func (c *Client) parseIssueCommentWebhook(jsonString []byte) (*git.Webhook, error) {
	issueComment := &IssueCommentWebhook{}
	if err := json.Unmarshal(jsonString, issueComment); err != nil {
		return nil, err
	}

	// Only handle creation
	if issueComment.Action != "created" {
		return nil, nil
	}

	// Get Pull Request info.s
	var pr *git.PullRequest
	if issueComment.Issue.PullRequest.URL != "" {
		prIDTokens := strings.Split(issueComment.Issue.PullRequest.URL, "/")
		prID, err := strconv.Atoi(prIDTokens[len(prIDTokens)-1])
		if err != nil {
			return nil, err
		}
		pr, err = c.GetPullRequest(prID)
		if err != nil {
			return nil, err
		}
	}

	// Get sender & author
	sender, author := c.getSenderAuthor(issueComment.Sender, issueComment.Comment.User)

	return &git.Webhook{EventType: git.EventTypeIssueComment, Repo: git.Repository{
		Name: issueComment.Repo.Name,
		URL:  issueComment.Repo.URL,
	},
		Sender:      *sender,
		RequestBody: string(jsonString),
		IssueComment: &git.IssueComment{
			Comment: git.Comment{
				Body:      issueComment.Comment.Body,
				CreatedAt: issueComment.Comment.CreatedAt,
			},
			Author: *author,
			Issue: git.Issue{
				PullRequest: pr,
			},
		}}, nil
}

func (c *Client) parsePullRequestReviewWebhook(jsonString []byte) (*git.Webhook, error) {
	review := &PullRequestReviewWebhook{}
	if err := json.Unmarshal(jsonString, review); err != nil {
		return nil, err
	}

	// Only handle creation
	if review.Action != "submitted" {
		return nil, nil
	}

	// Get sender & author
	sender, author := c.getSenderAuthor(review.Sender, review.Review.User)

	return &git.Webhook{EventType: git.EventTypePullRequestReview, Repo: git.Repository{
		Name: review.Repo.Name,
		URL:  review.Repo.URL,
	},
		Sender:      *sender,
		RequestBody: string(jsonString),
		IssueComment: &git.IssueComment{
			Comment: git.Comment{
				Body:      review.Review.Body,
				CreatedAt: review.Review.SubmittedAt,
			},
			Author:      *author,
			ReviewState: git.PullRequestReviewState(review.Review.State),
			Issue: git.Issue{
				PullRequest: convertPullRequestToShared(&review.PullRequest),
			},
		}}, nil
}

func (c *Client) parsePullRequestReviewCommentWebhook(jsonString []byte) (*git.Webhook, error) {
	reviewComment := &PullRequestReviewCommentWebhook{}
	if err := json.Unmarshal(jsonString, reviewComment); err != nil {
		return nil, err
	}

	// Only handle creation
	if reviewComment.Action != "created" {
		return nil, nil
	}

	// Get sender & author
	sender, author := c.getSenderAuthor(reviewComment.Sender, reviewComment.Comment.User)

	return &git.Webhook{EventType: git.EventTypePullRequestReviewComment, Repo: git.Repository{
		Name: reviewComment.Repo.Name,
		URL:  reviewComment.Repo.URL,
	},
		Sender:      *sender,
		RequestBody: string(jsonString),
		IssueComment: &git.IssueComment{
			Author: *author,
			Comment: git.Comment{
				Body:      reviewComment.Comment.Body,
				CreatedAt: reviewComment.Comment.CreatedAt,
			},
			Issue: git.Issue{
				PullRequest: convertPullRequestToShared(&reviewComment.PullRequest),
			},
		}}, nil
}

func (c *Client) getSenderAuthor(senderPre, authorPre User) (*git.User, *git.User) {
	// Get sender & email
	sender, err := c.GetUserInfo(senderPre.Name)
	if err != nil {
		sender = &git.User{Name: senderPre.Name, ID: senderPre.ID}
	}

	// Get author & email
	var author *git.User
	if sender.ID == authorPre.ID {
		author = sender
	} else {
		author, err = c.GetUserInfo(authorPre.Name)
		if err != nil {
			author = &git.User{Name: authorPre.Name, ID: authorPre.ID}
		}
	}

	return sender, author
}
