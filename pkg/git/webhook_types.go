package git

type EventType string

const (
	GitHubEventTypePullRequest  = EventType("pull_request")
	GitHubEventTypePush         = EventType("push")
	GitLabEventTypePush         = EventType("Push Hook")
	GitLabEventTypeMergeRequest = EventType("Merge Request Hook")
	GitLabEventTypeTagPush      = EventType("Tag Push Hook")
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
	Name string
	URL  string
}

type Sender struct {
	Name string
}

type Base struct {
	Ref string
}

type Head struct {
	Ref string
	Sha string
}
