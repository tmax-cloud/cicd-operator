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
}

type PullRequest struct {
	Title  string
	Sender string
	Url    string
}

type Repository struct {
	Name    string
	Owner   string
	Url     string
	Private bool
}
