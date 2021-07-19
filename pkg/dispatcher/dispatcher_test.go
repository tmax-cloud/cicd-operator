package dispatcher

import (
	"testing"

	"github.com/bmizerany/assert"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
)

func TestGeneratePull(t *testing.T) {
	pr := git.PullRequest{
		ID:     30,
		Author: git.User{Name: "Amy"},
		URL:    "https://api.github.com/repos/dev-yxzzzxh/test/pulls/6",
		Head: git.Head{
			Ref: "bugfix/first",
			Sha: "0kokpenadiugpowkqe0qlemaogor",
		},
	}

	pull := generatePull(pr)

	assert.Equal(t, 30, pull.ID)
	assert.Equal(t, "Amy", pull.Author.Name)
	assert.Equal(t, "bugfix/first", pull.Ref.String())
	assert.Equal(t, "0kokpenadiugpowkqe0qlemaogor", pull.Sha)
}

func TestGeneratePulls(t *testing.T) {
	prs := []git.PullRequest{
		{
			ID:     30,
			Author: git.User{Name: "Amy"},
			URL:    "https://api.github.com/repos/dev-yxzzzxh/test/pulls/6",
			Head: git.Head{
				Ref: "bugfix/first",
				Sha: "0kokpenadiugpowkqe0qlemaogor",
			},
		},
	}

	pulls := generatePulls(prs)

	assert.Equal(t, 30, pulls[0].ID)
	assert.Equal(t, "Amy", pulls[0].Author.Name)
	assert.Equal(t, "bugfix/first", pulls[0].Ref.String())
	assert.Equal(t, "0kokpenadiugpowkqe0qlemaogor", pulls[0].Sha)
}
