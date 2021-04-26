package gitlab

// UserInfo is a body of user get API
type UserInfo struct {
	ID          int    `json:"id"`
	UserName    string `json:"username"`
	PublicEmail string `json:"public_email"`
	Email       string `json:"email"`
}

// UserPermission is a user's permission on a repository
type UserPermission struct {
	AccessLevel int `json:"access_level"`
}

// CommitStatusRequest is an API body for setting commits' status
type CommitStatusRequest struct {
	State       string `json:"state"`
	TargetURL   string `json:"target_url"`
	Description string `json:"description"`
	Context     string `json:"context"`
}

// CommitStatusResponse is a response body of getting commit status
type CommitStatusResponse struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// CommentBody is a body structure for creating new comment
type CommentBody struct {
	Body string `json:"body"`
}

// UpdateMergeRequest is a struct to update a merge request
type UpdateMergeRequest struct {
	AddLabels    string `json:"add_labels"`
	RemoveLabels string `json:"remove_labels"`
}

// MergeRequest is a body struct of a merge request
type MergeRequest struct {
	ID     int    `json:"iid"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Author struct {
		ID       int    `json:"id"`
		UserName string `json:"username"`
	} `json:"author"`
	WebURL       string   `json:"web_url"`
	TargetBranch string   `json:"target_branch"`
	SourceBranch string   `json:"source_branch"`
	SHA          string   `json:"sha"`
	Labels       []string `json:"labels"`
	HasConflicts bool     `json:"has_conflicts"`
}

// BranchResponse is a respond struct for branch request
type BranchResponse struct {
	Name   string `json:"name"`
	Commit struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
}
