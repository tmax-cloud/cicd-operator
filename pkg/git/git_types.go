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

package git

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// EventType is a type of webhook event
type EventType string

// PullRequestState is a state of the pull request
type PullRequestState string

// PullRequestAction is an action of the pull request event
type PullRequestAction string

// PullRequestReviewState is a state of the pr's review
type PullRequestReviewState string

// Event Types
const (
	EventTypePullRequest              = EventType("pull_request")
	EventTypePush                     = EventType("push")
	EventTypeIssueComment             = EventType("issue_comment")
	EventTypePullRequestReview        = EventType("pull_request_review")
	EventTypePullRequestReviewComment = EventType("pull_request_review_comment")
)

// Pull Request states
const (
	PullRequestStateOpen   = PullRequestState("open")
	PullRequestStateClosed = PullRequestState("closed")
)

// Pull Request actions
const (
	PullRequestActionReOpen      = PullRequestAction("reopened")
	PullRequestActionOpen        = PullRequestAction("opened")
	PullRequestActionClose       = PullRequestAction("closed")
	PullRequestActionSynchronize = PullRequestAction("synchronize")
	PullRequestActionLabeled     = PullRequestAction("labeled")
	PullRequestActionUnlabeled   = PullRequestAction("unlabeled")
)

// Pull Request review state
const (
	PullRequestReviewStateApproved   = PullRequestReviewState("approved")
	PullRequestReviewStateUnapproved = PullRequestReviewState("changes_requested")
)

// Webhook is a common structure for git webhooks
// github-specific or gitlab-specific webhook bodies are converted to this structure before being consumed
type Webhook struct {
	EventType EventType
	Repo      Repository
	Sender    User

	Push         *Push
	PullRequest  *PullRequest
	IssueComment *IssueComment
}

// Push is a common structure for push events
type Push struct {
	Ref string
	Sha string
}

// PullRequest is a common structure for pull request events
type PullRequest struct {
	ID        int
	Title     string
	State     PullRequestState
	Action    PullRequestAction
	Author    User
	URL       string
	Base      Base
	Head      Head
	Labels    []IssueLabel
	Mergeable bool

	// LabelChanged
	LabelChanged []IssueLabel
}

// Diff is a diff between commits or of a pull-request
type Diff struct {
	Changes []Change
}

// Commit is a commit structure
type Commit struct {
	SHA       string
	Message   string
	Author    User
	Committer User
}

// Change is a diff of a changed file
type Change struct {
	Filename    string
	OldFilename string
	Additions   int
	Deletions   int
	Changes     int // Changes should be equal to the addition of Additions and Deletions
}

// IssueComment is a common structure for issue comment
type IssueComment struct {
	Comment     Comment
	Issue       Issue
	Author      User
	ReviewState PullRequestReviewState
}

// IssueLabel is a label struct
type IssueLabel struct {
	Name string
}

// Comment is a comment body
type Comment struct {
	Body string

	CreatedAt *metav1.Time
}

// Issue is an issue related to the Comment
type Issue struct {
	PullRequest *PullRequest
}

// Repository is a repository of the git
type Repository struct {
	Name string
	URL  string
}

// User is who triggered the event
type User struct {
	ID    int
	Name  string
	Email string
}

// Base is a reference for base commit
type Base struct {
	Ref string
	Sha string
}

// Head is a reference for head commit
type Head struct {
	Ref string
	Sha string
}

// WebhookEntry is a body of registered webhook list
type WebhookEntry struct {
	ID  int
	URL string
}

// CommitStatus is a commit status body
type CommitStatus struct {
	Context     string
	State       CommitStatusState
	Description string
	TargetURL   string
}

// Branch is a branch info
type Branch struct {
	Name     string
	CommitID string
}
