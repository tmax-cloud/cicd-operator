package gitlab

// Webhook is a gitlab-specific webhook body
// Should contain json tag for each field, to be unmarshalled properly
type MergeRequestWebhook struct {
	Kind            string          `json:"kind"`
	User            User            `json:"user"`
	ObjectAttribute ObjectAttribute `json:"object_attributes"`
	Project         Project         `json:"project"`
}

// Structure for push webhook event
type PushWebhook struct {
	Kind     string  `json:"object_kind"`
	Ref      string  `json:"ref"`
	Project  Project `json:"project"`
	UserName string  `json:"user_name"`
	UserId   int     `json:"user_id"`
	Sha      string  `json:"after"`
}

type Project struct {
	Name   string `json:"path_with_namespace"`
	WebUrl string `json:"web_url"`
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type ObjectAttribute struct {
	Title      string     `json:"title"`
	ID         int        `json:"id"`
	BaseRef    string     `json:"target_branch"`
	HeadRef    string     `json:"source_branch"`
	LastCommit LastCommit `json:"last_commit"`
	State      string     `json:"state"`
	Action     string     `json:"action"`
}
type LastCommit struct {
	Sha string `json:"id"`
}

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

type WebhookEntry struct {
	Id  int    `json:"id"`
	Url string `json:"url"`
}
