package gitlab

// Webhook is a gitlab-specific webhook body
// Should contain json tag for each field, to be unmarshalled properly
type MergeRequestWebhook struct {
	Kind            string          `json:"kind"`
	Sender          Sender          `json:"user"`
	ObjectAttribute ObjectAttribute `json:"object_attributes"`
	Repo            Repo            `json:"repository"`
}

// Structure for push webhook event
type PushWebhook struct {
	Kind string `json:"object_kind"`
	Ref  string `json:"ref"`
	Repo Repo   `json:"repository"`
	User string `json:"user_name"`
	Sha  string `json:"checkout_sha"`
}

// Repo structure for webhook event
type Repo struct {
	Name    string `json:"name"`
	Htmlurl string `json:"url"`
}

type Sender struct {
	ID string `json:"username"`
}

type ObjectAttribute struct {
	Title      string     `json:"title"`
	ID         int        `json:"id"`
	BaseRef    string     `json:"target_branch"`
	HeadRef    string     `json:"source_branch"`
	LastCommit LastCommit `json:"last_commit"`
}
type LastCommit struct {
	Sha string `json:"id"`
}
