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
