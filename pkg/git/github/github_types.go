package github

// UserInfo is a body of user get API
type UserInfo struct {
	ID       int    `json:"id"`
	UserName string `json:"login"`
	Email    string `json:"email"`
}

// UserPermission is a user's permission on a repository
type UserPermission struct {
	Permission string `json:"permission"`
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
	Context string `json:"context"`
	State   string `json:"state"`
}

// CommentBody is a body structure for creating new comment
type CommentBody struct {
	Body string `json:"body"`
}

// LabelBody is a body structure for setting a label to issues/prs
type LabelBody struct {
	Name string `json:"name"`
}
