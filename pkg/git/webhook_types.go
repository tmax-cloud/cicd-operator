package git

type EventType string

const (
	EventTypePullRequest = EventType("pull_request")
	EventTypePush        = EventType("push")
)

// Webhook is a common structure for git webhooks
// github-specific or gitlab-specific webhook bodies are converted to this structure before being consumed
type Webhook struct {
	EventType EventType
	Repo      Repository

	Push        *Push
	PullRequest *PullRequest
}

type Push struct {
	Pusher string
	Ref    string
	Sha    string
}

type PullRequest struct {
	ID     int
	Title  string
	Sender Sender
	URL    string
	Base   Base
	Head   Head
}

type Repository struct {
	Name    string
	Owner   string
	URL     string
	Private bool
}

type Sender struct {
	Name string
	Link string
}

type Base struct {
	Ref string
	Sha string
}

type Head struct {
	Ref string
	Sha string
}
