package git

// EventType is a type of webhook event
type EventType string

// PullRequestState is a state of the pull request
type PullRequestState string

// PullRequestAction is an action of the pull request event
type PullRequestAction string

// Event Types
const (
	EventTypePullRequest = EventType("pull_request")
	EventTypePush        = EventType("push")
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
)

// Webhook is a common structure for git webhooks
// github-specific or gitlab-specific webhook bodies are converted to this structure before being consumed
type Webhook struct {
	EventType EventType
	Repo      Repository

	Push        *Push
	PullRequest *PullRequest
	Action      string
}

// Push is a common structure for push events
type Push struct {
	Sender Sender
	Ref    string
	Sha    string
}

// PullRequest is a common structure for pull request events
type PullRequest struct {
	ID     int
	Title  string
	State  PullRequestState
	Action PullRequestAction
	Sender Sender
	URL    string
	Base   Base
	Head   Head
}

// Repository is a repository of the git
type Repository struct {
	Name string
	URL  string
}

// Sender is who triggered the event
type Sender struct {
	ID    int
	Name  string
	Email string
}

// Base is a reference for base commit
type Base struct {
	Ref string
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
