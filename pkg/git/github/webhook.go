package github

// Webhook is a github-specific webhook body
// Should contain json tag for each field, to be unmarshalled properly
type PullRequestWebhook struct {
	Sender struct {
		ID string `json:"login"`
	} `json:"sender"`
	PullRequest struct {
		Title string `json:"title"`
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
}

// Repo structure for webhook event
type Repo struct {
	Name    string `json:"name"`
	Htmlurl string `json:"html_url"`
	Owner   struct {
		ID string `json:"login"`
	} `json:"owner"`
	Private bool `json:"private"`
}
