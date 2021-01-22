package git

type EventType string
type PullRequestState string
type PullRequestAction string

const (
	EventTypePullRequest = EventType("pull_request")
	EventTypePush        = EventType("push")

	PullRequestStateOpen   = PullRequestState("open")
	PullRequestStateClosed = PullRequestState("closed")

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

type Push struct {
	Sender Sender
	Ref    string
	Sha    string
}

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

type Repository struct {
	Name string
	URL  string
}

type Sender struct {
	ID    int
	Name  string
	Email string
}

type Base struct {
	Ref string
}

type Head struct {
	Ref string
	Sha string
}

type WebhookEntry struct {
	Id  int
	Url string
}
