package gitlab

// MergeRequestWebhook is a gitlab-specific merge-request event webhook body
type MergeRequestWebhook struct {
	Kind            string          `json:"kind"`
	User            User            `json:"user"`
	ObjectAttribute ObjectAttribute `json:"object_attributes"`
	Project         Project         `json:"project"`
}

// PushWebhook is a gitlab-specific push event webhook body
type PushWebhook struct {
	Kind     string  `json:"object_kind"`
	Ref      string  `json:"ref"`
	Project  Project `json:"project"`
	UserName string  `json:"user_name"`
	UserID   int     `json:"user_id"`
	Sha      string  `json:"after"`
}

// Project is a name/url for the repository
type Project struct {
	Name   string `json:"path_with_namespace"`
	WebURL string `json:"web_url"`
}

// User is a user who triggered merge request event
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ObjectAttribute is a core body for pull request webhook
type ObjectAttribute struct {
	Title      string     `json:"title"`
	ID         int        `json:"id"`
	BaseRef    string     `json:"target_branch"`
	HeadRef    string     `json:"source_branch"`
	LastCommit LastCommit `json:"last_commit"`
	State      string     `json:"state"`
	Action     string     `json:"action"`
}

// LastCommit is a commit hash for the last commit of the pull request
type LastCommit struct {
	Sha string `json:"id"`
}

// RegistrationWebhookBody is a body for requesting webhook registration for the remote git server
type RegistrationWebhookBody struct {
	EnableSSLVerification   bool   `json:"enable_ssl_verification"`
	ConfidentialIssueEvents bool   `json:"confidential_issues_events"`
	ConfidentialNoteEvents  bool   `json:"confidential_note_events"`
	DeploymentEvents        bool   `json:"deployment_events"`
	ID                      string `json:"id"`
	IssueEvents             bool   `json:"issues_events"`
	JobEvents               bool   `json:"job_events"`
	MergeRequestEvents      bool   `json:"merge_requests_events"`
	NoteEvents              bool   `json:"note_events"`
	PipeLineEvents          bool   `json:"pipeline_events"`
	PushEvents              bool   `json:"push_events"`
	TagPushEvents           bool   `json:"tag_push_events"`
	WikiPageEvents          bool   `json:"wiki_page_events"`
	URL                     string `json:"url"`
	Token                   string `json:"token"`
}

// WebhookEntry is a body of list of registered webhooks
type WebhookEntry struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}
