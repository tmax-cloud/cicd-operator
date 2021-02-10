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
	sender := git.User{Name: data.User.Name, Email: data.User.Email}
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
		action = git.PullRequestActionSynchronize
	}
	state := git.PullRequestState(data.ObjectAttribute.State)
	switch string(state) {
	case "opened":
		state = git.PullRequestStateOpen
	case "closed":
		state = git.PullRequestStateClosed
	}
	pullRequest := git.PullRequest{ID: data.ObjectAttribute.ID, Title: data.ObjectAttribute.Title, Sender: sender, URL: data.Project.WebURL, Base: base, Head: head, State: state, Action: action}
	return &git.Webhook{EventType: git.EventTypePullRequest, Repo: repo, PullRequest: &pullRequest}, nil
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
	push := git.Push{Sender: git.User{Name: data.UserName, ID: data.UserID}, Ref: data.Ref, Sha: data.Sha}

	// Get sender email
	userInfo, err := c.GetUserInfo(strconv.Itoa(data.UserID))
	if err == nil {
		push.Sender.Email = userInfo.Email
	}

	return &git.Webhook{EventType: git.EventTypePush, Repo: repo, Push: &push}, nil
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

	// Get Merge Request user info
	var pr *git.PullRequest
	if data.MergeRequest.TargetBranch != "" {
		mrAuthor, err := c.GetUserInfo(strconv.Itoa(data.MergeRequest.AuthorID))
		if err != nil {
			mrAuthor = &git.User{ID: data.MergeRequest.AuthorID}
		}
		pr = &git.PullRequest{
			ID:     data.MergeRequest.ID,
			Title:  data.MergeRequest.Title,
			State:  mrState,
			Sender: *mrAuthor,
			URL:    data.MergeRequest.URL,
			Base: git.Base{
				Ref: data.MergeRequest.TargetBranch,
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
	}, IssueComment: &git.IssueComment{
		Comment: git.Comment{
			Body:      data.ObjectAttributes.Note,
			CreatedAt: &metav1.Time{Time: data.ObjectAttributes.CreatedAt.Time},
		},
		Issue: git.Issue{
			PullRequest: pr,
		},
		Sender: git.User{
			ID:    data.ObjectAttributes.AuthorID,
			Name:  data.User.Name,
			Email: data.User.Email,
		},
	}}, nil
}
