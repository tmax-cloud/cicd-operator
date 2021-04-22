package git

import "fmt"

// UnauthorizedError is an error struct for git clients
type UnauthorizedError struct {
	User string
	Repo string
}

// Error returns error string
func (e *UnauthorizedError) Error() string {
	return fmt.Sprintf("%s is not authorized for %s", e.User, e.Repo)
}
