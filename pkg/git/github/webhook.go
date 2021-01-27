package github

// PullRequestWebhook is a github-specific pull-request event webhook body
type PullRequestWebhook struct {
	Action string `json:"action"`
	Number int    `json:"number"`
	Sender Sender `json:"sender"`

	PullRequest struct {
		Title string `json:"title"`
		ID    int    `json:"id"`
		State string `json:"state"`
		Head  Head   `json:"head"`
		Base  Base   `json:"base"`
	} `json:"pull_request"`

	Repo Repo `json:"repository"`
}

// PushWebhook is a github-specific push event webhook body
type PushWebhook struct {
	Ref    string `json:"ref"`
	Repo   Repo   `json:"repository"`
	Sender Sender `json:"sender"`
	Sha    string `json:"after"`
}

// Repo structure for webhook event
type Repo struct {
	Name    string `json:"full_name"`
	HTMLURL string `json:"html_url"`
	Owner   struct {
		ID string `json:"login"`
	} `json:"owner"`
	Private bool `json:"private"`
}

// Sender is a sender of the event
type Sender struct {
	Name string `json:"login"`
	ID   int    `json:"id"`
}

// Head is a reference of the head branch
type Head struct {
	Ref string `json:"ref"`
	Sha string `json:"sha"`
}

// Base is a reference of the base branch
type Base struct {
	Ref string `json:"ref"`
}

// RegistrationWebhookBody is a request body for registering webhook to remote git server
type RegistrationWebhookBody struct {
	Name   string                        `json:"name"`
	Active bool                          `json:"active"`
	Events []string                      `json:"events"`
	Config RegistrationWebhookBodyConfig `json:"config"`
}

// RegistrationWebhookBodyConfig is a config for the webhook
type RegistrationWebhookBodyConfig struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	InsecureSsl string `json:"insecure_ssl"`
	Secret      string `json:"secret"`
}

// WebhookEntry is a body of list of registered webhooks
type WebhookEntry struct {
	ID     int `json:"id"`
	Config struct {
		URL string `json:"url"`
	} `json:"config"`
}
