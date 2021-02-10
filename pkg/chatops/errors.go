package chatops

import "fmt"

type unauthorizedError struct {
	user string
	repo string
}

func (e *unauthorizedError) Error() string {
	return fmt.Sprintf("%s is not authorized for %s", e.user, e.repo)
}
