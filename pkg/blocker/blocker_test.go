package blocker

import (
	"github.com/bmizerany/assert"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"testing"
)

func TestMergePool_Add(t *testing.T) {
	pool := NewMergePool()

	pool.Add(&PullRequest{
		PullRequest: git.PullRequest{
			ID: 23,
		},
		BlockerStatus:      git.CommitStatusStatePending,
		BlockerDescription: defaultBlockerMessage,
	})

	assert.Equal(t, 1, len(pool[git.CommitStatusStatePending]), "Pending length")
	assert.Equal(t, 0, len(pool[git.CommitStatusStateSuccess]), "Success length")
	assert.Equal(t, defaultBlockerMessage, pool[git.CommitStatusStatePending][23].BlockerDescription)

	pool.Add(&PullRequest{
		PullRequest: git.PullRequest{
			ID: 25,
		},
		BlockerStatus:      git.CommitStatusStateSuccess,
		BlockerDescription: "test message",
	})

	assert.Equal(t, 1, len(pool[git.CommitStatusStatePending]), "Pending length")
	assert.Equal(t, 1, len(pool[git.CommitStatusStateSuccess]), "Success length")
	assert.Equal(t, "test message", pool[git.CommitStatusStateSuccess][25].BlockerDescription)
}

func TestMergePool_Search(t *testing.T) {
	pool := NewMergePool()

	pool[git.CommitStatusStatePending][23] = &PullRequest{
		PullRequest: git.PullRequest{
			ID: 23,
		},
		BlockerStatus:      git.CommitStatusStatePending,
		BlockerDescription: defaultBlockerMessage,
	}

	pool[git.CommitStatusStatePending][25] = &PullRequest{
		PullRequest: git.PullRequest{
			ID: 25,
		},
		BlockerStatus:      git.CommitStatusStateSuccess,
		BlockerDescription: "test message",
	}

	found := pool.Search(23)
	assert.Equal(t, defaultBlockerMessage, found.BlockerDescription, "Found description")

	found = pool.Search(25)
	assert.Equal(t, "test message", found.BlockerDescription, "Found description")

	found = pool.Search(27)
	assert.Equal(t, true, found == nil, "Not found")
}

func TestMergePool_Delete(t *testing.T) {
	pool := NewMergePool()

	pool[git.CommitStatusStatePending][23] = &PullRequest{
		PullRequest: git.PullRequest{
			ID: 23,
		},
		BlockerStatus:      git.CommitStatusStatePending,
		BlockerDescription: defaultBlockerMessage,
	}
	assert.Equal(t, 1, len(pool[git.CommitStatusStatePending]))

	pool.Delete(23)

	assert.Equal(t, 0, len(pool[git.CommitStatusStatePending]))
}
