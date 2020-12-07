package github

// Webhook is a github-specific webhook body
// Should contain json tag for each field, to be unmarshalled properly
type PullRequestWebhook struct {
	Sender Sender `json:"sender"`

	PullRequest struct {
		Title string `json:"title"`
		ID    int    `json:"id"`
		Head  Head   `json:"head"`
		Base  Base   `json:"base"`
	} `json:"pull_request"`

	Repo Repo `json:"repository"`
}

// Structure for push webhook event
type PushWebhook struct {
	Ref    string `json:"ref"`
	Repo   Repo   `json:"repository"`
	Pusher struct {
		Name string `json:"name"`
	} `json:"pusher"`
	Sha4Push string `json:"after"`
	Sha4Tag  string `json:"before"`
}

// Repo structure for webhook event
type Repo struct {
	Name    string `json:"full_name"`
	Htmlurl string `json:"html_url"`
	Owner   struct {
		ID string `json:"login"`
	} `json:"owner"`
	Private bool `json:"private"`
}

type Sender struct {
	ID string `json:"login"`
}

type Head struct {
	Ref string `json:"ref"`
	Sha string `json:"sha"`
}

type Base struct {
	Ref string `json:"ref"`
}

type RegistrationWebhookBody struct {
	Name   string                        `json:"name"`
	Active bool                          `json:"active"`
	Events []string                      `json:"events"`
	Config RegistrationWebhookBodyConfig `json:"config"`
}

type RegistrationWebhookBodyConfig struct {
	Url         string `json:"url"`
	ContentType string `json:"content_type"`
	InsecureSsl string `json:"insecure_ssl"`
	Secret      string `json:"secret"`
}
