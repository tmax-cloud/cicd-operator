package gitlab

// MergeRequestWebhook is a gitlab-specific merge-request event webhook body
type MergeRequestWebhook struct {
	Kind            string `json:"kind"`
	User            User   `json:"user"`
	ObjectAttribute struct {
		Title      string `json:"title"`
		ID         int    `json:"id"`
		BaseRef    string `json:"target_branch"`
		HeadRef    string `json:"source_branch"`
		LastCommit struct {
			Sha string `json:"id"`
		} `json:"last_commit"`
		State  string `json:"state"`
		Action string `json:"action"`
	} `json:"object_attributes"`
	Project Project `json:"project"`
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

// NoteHook is a gitlab-specific issue comment webhook body
type NoteHook struct {
	User             User    `json:"user"`
	Project          Project `json:"project"`
	ObjectAttributes struct {
		Note      string     `json:"note"`
		AuthorID  int        `json:"author_id"`
		CreatedAt gitlabTime `json:"created_at"`
		UpdatedAt gitlabTime `json:"updated_at"`
	} `json:"object_attributes"`
	MergeRequest struct {
		ID           int    `json:"id"`
		Title        string `json:"title"`
		State        string `json:"state"`
		URL          string `json:"url"`
		AuthorID     int    `json:"author_id"`
		SourceBranch string `json:"source_branch"`
		TargetBranch string `json:"target_branch"`
		LastCommit   struct {
			ID string `json:"id"`
		} `json:"last_commit"`
	} `json:"merge_request"`
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
