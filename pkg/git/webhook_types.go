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
<<<<<<< HEAD
	Name string
	URL  string
=======
	Name    string
	Owner   string
	URL     string
	Private bool
>>>>>>> 21401d5ca1a5ec607545a8ab68a2e18f82364be2
}

type Sender struct {
	Name string
<<<<<<< HEAD
=======
	Link string
>>>>>>> 21401d5ca1a5ec607545a8ab68a2e18f82364be2
}

type Base struct {
	Ref string
}

type Head struct {
	Ref string
	Sha string
}
