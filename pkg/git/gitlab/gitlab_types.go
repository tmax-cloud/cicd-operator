package gitlab

// UserInfo is a body of user get API
type UserInfo struct {
	ID          int    `json:"id"`
	UserName    string `json:"username"`
	PublicEmail string `json:"public_email"`
	Email       string `json:"email"`
}

// CommitStatusBody is an API body for setting commits' status
type CommitStatusBody struct {
	State       string `json:"state"`
	TargetURL   string `json:"target_url"`
	Description string `json:"description"`
	Context     string `json:"context"`
}
