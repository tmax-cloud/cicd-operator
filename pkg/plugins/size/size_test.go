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

package size

import (
	"testing"

	"github.com/stretchr/testify/require"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	gitfake "github.com/tmax-cloud/cicd-operator/pkg/git/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testRepo = "tmax-cloud/cicd-operator"
	testPRID = 1
)

type sizeTestCase struct {
	labels            []git.IssueLabel
	additions         int
	deletions         int
	errorOccurs       bool
	errorMessage      string
	expectedLabelsLen int
	expectedLabel     string
}

func TestSize_Handle(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))

	ic := &cicdv1.IntegrationConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: cicdv1.IntegrationConfigSpec{
			Git: cicdv1.GitConfig{
				Type:       cicdv1.GitTypeFake,
				Repository: testRepo,
			},
		},
	}

	fakeCli := fake.NewClientBuilder().WithScheme(s).WithObjects(ic).Build()
	size := Size{Client: fakeCli}

	tc := map[string]sizeTestCase{
		"XS": {
			labels:            nil,
			additions:         1,
			deletions:         2,
			expectedLabelsLen: 1,
			expectedLabel:     "size/XS",
		},
		"L": {
			labels:            nil,
			additions:         100,
			deletions:         101,
			expectedLabelsLen: 1,
			expectedLabel:     "size/L",
		},
		"changeLabel": {
			labels:            []git.IssueLabel{{Name: "size/XS"}},
			additions:         100,
			deletions:         101,
			expectedLabelsLen: 1,
			expectedLabel:     "size/L",
		},
		"changeLabel2": {
			labels:            []git.IssueLabel{{Name: "size/XS"}, {Name: "approved"}},
			additions:         100,
			deletions:         101,
			expectedLabelsLen: 2,
			expectedLabel:     "size/L",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// Set fake git
			gitfake.Repos = map[string]*gitfake.Repo{
				testRepo: {
					PullRequests: map[int]*git.PullRequest{
						testPRID: {Labels: c.labels},
					},
					PullRequestDiffs: map[int]*git.Diff{
						testPRID: {
							Changes: []git.Change{{
								Additions: c.additions,
								Deletions: c.deletions,
								Changes:   c.additions + c.deletions,
							}},
						},
					},
				},
			}

			wh := &git.Webhook{
				EventType: git.EventTypePullRequest,
				PullRequest: &git.PullRequest{
					ID:     testPRID,
					Action: git.PullRequestActionOpen,
					Labels: c.labels,
				},
			}
			err := size.Handle(wh, ic)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Len(t, gitfake.Repos[testRepo].PullRequests[testPRID].Labels, c.expectedLabelsLen)
				require.Contains(t, gitfake.Repos[testRepo].PullRequests[testPRID].Labels, git.IssueLabel{Name: c.expectedLabel})
				for _, l := range gitfake.Repos[testRepo].PullRequests[testPRID].Labels {
					if l.Name != c.expectedLabel {
						require.NotContains(t, l.Name, labelPrefix)
					}
				}
			}
		})
	}
}
