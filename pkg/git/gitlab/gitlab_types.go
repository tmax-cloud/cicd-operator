package gitlab

type UserInfo struct {
	ID          int    `json:"id"`
	UserName    string `json:"username"`
	PublicEmail string `json:"public_email"`
	Email       string `json:"email"`
}
