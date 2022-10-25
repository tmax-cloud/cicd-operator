/*
 Copyright 2021 The CI/CD Operator Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

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
