package github

import (
	"encoding/json"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"strconv"
	"strings"
)

func (c *Client) parsePullRequestWebhook(jsonString []byte) (*git.Webhook, error) {
	var data PullRequestWebhook

	if err := json.Unmarshal(jsonString, &data); err != nil {
		return nil, err
	}

	// Get sender email
	sender := git.User{Name: data.Sender.Name, ID: data.Sender.ID}
	userInfo, err := c.GetUserInfo(data.Sender.Name)
	if err == nil {
		sender.Email = userInfo.Email
	}

	var labels []git.IssueLabel
	for _, l := range data.PullRequest.Labels {
		labels = append(labels, git.IssueLabel{Name: l.Name})
	}

	base := git.Base{Ref: data.PullRequest.Base.Ref, Sha: data.PullRequest.Base.Sha}
	head := git.Head{Ref: data.PullRequest.Head.Ref, Sha: data.PullRequest.Head.Sha}
	repo := git.Repository{Name: data.Repo.Name, URL: data.Repo.URL}
	pullRequest := git.PullRequest{ID: data.Number, Title: data.PullRequest.Title, Sender: sender, URL: data.Repo.URL, Base: base, Head: head, State: git.PullRequestState(data.PullRequest.State), Action: git.PullRequestAction(data.Action), Labels: labels}
	return &git.Webhook{EventType: git.EventTypePullRequest, Repo: repo, PullRequest: &pullRequest}, nil
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
	push := git.Push{Sender: git.User{Name: data.Sender.Name, ID: data.Sender.ID}, Ref: data.Ref, Sha: data.Sha}

	// Get sender email
	userInfo, err := c.GetUserInfo(data.Sender.Name)
	if err == nil {
		push.Sender.Email = userInfo.Email
	}

	return &git.Webhook{EventType: git.EventTypePush, Repo: repo, Push: &push}, nil
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

	// Get sender email
	sender, err := c.GetUserInfo(issueComment.Sender.Name)
	if err != nil {
		sender = &git.User{Name: issueComment.Sender.Name, ID: issueComment.Sender.ID}
	}

	return &git.Webhook{EventType: git.EventTypeIssueComment, Repo: git.Repository{
		Name: issueComment.Repo.Name,
		URL:  issueComment.Repo.URL,
	}, IssueComment: &git.IssueComment{
		Comment: git.Comment{
			Body:      issueComment.Comment.Body,
			CreatedAt: issueComment.Comment.CreatedAt,
		},
		Issue: git.Issue{
			PullRequest: pr,
		},
		Sender: *sender,
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

	// Get sender email
	sender, err := c.GetUserInfo(review.Sender.Name)
	if err != nil {
		sender = &git.User{Name: review.Sender.Name, ID: review.Sender.ID}
	}

	return &git.Webhook{EventType: git.EventTypePullRequestReview, Repo: git.Repository{
		Name: review.Repo.Name,
		URL:  review.Repo.URL,
	}, IssueComment: &git.IssueComment{
		Comment: git.Comment{
			Body:      review.Review.Body,
			CreatedAt: review.Review.SubmittedAt,
		},
		ReviewState: git.PullRequestReviewState(review.Review.State),
		Issue: git.Issue{
			PullRequest: convertPullRequestToShared(&review.PullRequest),
		},
		Sender: *sender,
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

	// Get sender email
	sender, err := c.GetUserInfo(reviewComment.Sender.Name)
	if err != nil {
		sender = &git.User{Name: reviewComment.Sender.Name, ID: reviewComment.Sender.ID}
	}

	return &git.Webhook{EventType: git.EventTypePullRequestReviewComment, Repo: git.Repository{
		Name: reviewComment.Repo.Name,
		URL:  reviewComment.Repo.URL,
	}, IssueComment: &git.IssueComment{
		Comment: git.Comment{
			Body:      reviewComment.Comment.Body,
			CreatedAt: reviewComment.Comment.CreatedAt,
		},
		Issue: git.Issue{
			PullRequest: convertPullRequestToShared(&reviewComment.PullRequest),
		},
		Sender: *sender,
	}}, nil
}
