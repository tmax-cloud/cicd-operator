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
	"context"
	"os"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/stretchr/testify/require"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	gitfake "github.com/tmax-cloud/cicd-operator/pkg/git/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	testICName      = "test-ic"
	testICNamespace = "default"
)

func TestSortPullRequestByID(t *testing.T) {
	prs := map[int]*PullRequest{
		13: {PullRequest: git.PullRequest{ID: 13}},
		6:  {PullRequest: git.PullRequest{ID: 6}},
		72: {PullRequest: git.PullRequest{ID: 72}},
	}

	sorted := sortPullRequestByID(prs)

	assert.Equal(t, 3, len(sorted), "Length")
	assert.Equal(t, 6, sorted[0].ID, "1st PR")
	assert.Equal(t, 13, sorted[1].ID, "2nd PR")
	assert.Equal(t, 72, sorted[2].ID, "3rd PR")
}

func TestBlocker_retestAndMergeOnePool(t *testing.T) {
	tc := map[string]struct {
		prs           []*PullRequest
		baseSHA       string
		existingBatch *Batch
		existingJob   *cicdv1.IntegrationJob

		expectedIJRefPulls   []cicdv1.IntegrationJobRefsPull
		expectedBatchCreated bool
		expectedPRMerged     bool
	}{
		"successful": {
			baseSHA: "22ccae53032027186ba739dfaa473ee61a82b298",
			prs: []*PullRequest{
				{
					PullRequest: git.PullRequest{
						ID:        12,
						Base:      git.Base{Ref: "master", Sha: "22ccae53032027186ba739dfaa473ee61a82b298"},
						Head:      git.Head{Ref: "newnew", Sha: "3196ccc37bcae94852079b04fcbfaf928341d6e9"},
						Mergeable: true,
						State:     git.PullRequestStateOpen,
					},
					BlockerStatus: git.CommitStatusStateSuccess,
					Statuses: map[string]git.CommitStatus{
						"test-1": {Context: "test-1", State: git.CommitStatusStateSuccess, Description: "Job is successful    BaseSHA:22ccae53032027186ba739dfaa473ee61a82b298"},
					},
				},
			},
			expectedPRMerged: true,
		},
		"baseUpdates": {
			baseSHA: "32cd89e8d07e37ab26d8c735090ae763884283db",
			prs: []*PullRequest{
				{
					PullRequest: git.PullRequest{
						ID:        12,
						Base:      git.Base{Ref: "master", Sha: "22ccae53032027186ba739dfaa473ee61a82b298"},
						Head:      git.Head{Ref: "newnew", Sha: "3196ccc37bcae94852079b04fcbfaf928341d6e9"},
						Mergeable: true,
						State:     git.PullRequestStateOpen,
					},
					BlockerStatus: git.CommitStatusStateSuccess,
					Statuses: map[string]git.CommitStatus{
						"test-1": {Context: "test-1", State: git.CommitStatusStateSuccess, Description: "Job is successful    BaseSHA:22ccae53032027186ba739dfaa473ee61a82b298"},
					},
				},
			},
			expectedIJRefPulls: []cicdv1.IntegrationJobRefsPull{
				{ID: 12, Ref: "newnew", Sha: "3196ccc37bcae94852079b04fcbfaf928341d6e9", Author: cicdv1.IntegrationJobRefsPullAuthor{}},
			},
			expectedBatchCreated: true,
		},
		"multiplePRsRetest": {
			baseSHA: "32cd89e8d07e37ab26d8c735090ae763884283db",
			prs: []*PullRequest{
				{
					PullRequest: git.PullRequest{
						ID:        12,
						Base:      git.Base{Ref: "master", Sha: "22ccae53032027186ba739dfaa473ee61a82b298"},
						Head:      git.Head{Ref: "fix/1", Sha: "3196ccc37bcae94852079b04fcbfaf928341d6e9"},
						Mergeable: true,
						State:     git.PullRequestStateOpen,
					},
					BlockerStatus: git.CommitStatusStateSuccess,
					Statuses: map[string]git.CommitStatus{
						"test-1": {Context: "test-1", State: git.CommitStatusStateSuccess, Description: "Job is successful    BaseSHA:22ccae53032027186ba739dfaa473ee61a82b298"},
					},
				},
				{
					PullRequest: git.PullRequest{
						ID:        13,
						Base:      git.Base{Ref: "master", Sha: "22ccae53032027186ba739dfaa473ee61a82b298"},
						Head:      git.Head{Ref: "fix/2", Sha: "3bede531bd0bbe8d3735f2642193fb33800149e0"},
						Mergeable: true,
						State:     git.PullRequestStateOpen,
					},
					BlockerStatus: git.CommitStatusStateSuccess,
					Statuses: map[string]git.CommitStatus{
						"test-1": {Context: "test-1", State: git.CommitStatusStateSuccess, Description: "Job is successful    BaseSHA:22ccae53032027186ba739dfaa473ee61a82b298"},
					},
				},
			},
			expectedIJRefPulls: []cicdv1.IntegrationJobRefsPull{
				{ID: 12, Ref: "fix/1", Sha: "3196ccc37bcae94852079b04fcbfaf928341d6e9", Author: cicdv1.IntegrationJobRefsPullAuthor{}},
				{ID: 13, Ref: "fix/2", Sha: "3bede531bd0bbe8d3735f2642193fb33800149e0", Author: cicdv1.IntegrationJobRefsPullAuthor{}},
			},
			expectedBatchCreated: true,
		},
		"batchSuccessful": {
			baseSHA: "32cd89e8d07e37ab26d8c735090ae763884283db",
			prs: []*PullRequest{
				{
					PullRequest: git.PullRequest{
						ID:        12,
						Base:      git.Base{Ref: "master", Sha: "22ccae53032027186ba739dfaa473ee61a82b298"},
						Head:      git.Head{Ref: "fix/1", Sha: "3196ccc37bcae94852079b04fcbfaf928341d6e9"},
						Mergeable: true,
						State:     git.PullRequestStateOpen,
					},
					BlockerStatus: git.CommitStatusStateSuccess,
					Statuses: map[string]git.CommitStatus{
						"test-1": {Context: "test-1", State: git.CommitStatusStateSuccess, Description: "Job is successful    BaseSHA:22ccae53032027186ba739dfaa473ee61a82b298"},
					},
				},
				{
					PullRequest: git.PullRequest{
						ID:        13,
						Base:      git.Base{Ref: "master", Sha: "22ccae53032027186ba739dfaa473ee61a82b298"},
						Head:      git.Head{Ref: "fix/2", Sha: "3bede531bd0bbe8d3735f2642193fb33800149e0"},
						Mergeable: true,
						State:     git.PullRequestStateOpen,
					},
					BlockerStatus: git.CommitStatusStateSuccess,
					Statuses: map[string]git.CommitStatus{
						"test-1": {Context: "test-1", State: git.CommitStatusStateSuccess, Description: "Job is successful    BaseSHA:22ccae53032027186ba739dfaa473ee61a82b298"},
					},
				},
			},
			existingBatch: &Batch{
				PRs: []*PullRequest{
					{
						PullRequest: git.PullRequest{
							ID:        12,
							Base:      git.Base{Ref: "master", Sha: "22ccae53032027186ba739dfaa473ee61a82b298"},
							Head:      git.Head{Ref: "fix/1", Sha: "3196ccc37bcae94852079b04fcbfaf928341d6e9"},
							Mergeable: true,
							State:     git.PullRequestStateOpen,
						},
						BlockerStatus: git.CommitStatusStateSuccess,
						Statuses: map[string]git.CommitStatus{
							"test-1": {Context: "test-1", State: git.CommitStatusStateSuccess, Description: "Job is successful    BaseSHA:22ccae53032027186ba739dfaa473ee61a82b298"},
						},
					},
					{
						PullRequest: git.PullRequest{
							ID:        13,
							Base:      git.Base{Ref: "master", Sha: "22ccae53032027186ba739dfaa473ee61a82b298"},
							Head:      git.Head{Ref: "fix/2", Sha: "3bede531bd0bbe8d3735f2642193fb33800149e0"},
							Mergeable: true,
							State:     git.PullRequestStateOpen,
						},
						BlockerStatus: git.CommitStatusStateSuccess,
						Statuses: map[string]git.CommitStatus{
							"test-1": {Context: "test-1", State: git.CommitStatusStateSuccess, Description: "Job is successful    BaseSHA:22ccae53032027186ba739dfaa473ee61a82b298"},
						},
					},
				},
				Job: types.NamespacedName{Name: "batch-test", Namespace: "default"},
			},
			existingJob: &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{Name: "batch-test", Namespace: "default"},
				Spec:       cicdv1.IntegrationJobSpec{},
				Status:     cicdv1.IntegrationJobStatus{State: cicdv1.IntegrationJobStateCompleted},
			},
			expectedPRMerged: true,
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			// Init
			ic, cli := mergeTestConfig()
			b := New(cli)
			gitfake.Repos = map[string]*gitfake.Repo{
				ic.Spec.Git.Repository: {PullRequests: map[int]*git.PullRequest{}, Commits: map[string][]git.Commit{}},
			}
			gitfake.Branches = map[string]*git.Branch{
				"master": {CommitID: c.baseSHA},
			}
			pool := NewPRPool(testICNamespace, testICName)
			for _, pr := range c.prs {
				pool.PullRequests[pr.ID] = pr
				pool.MergePool.Add(pr)
				gitfake.Repos[ic.Spec.Git.Repository].PullRequests[pr.ID] = &pr.PullRequest
			}
			pool.CurrentBatch = c.existingBatch

			if c.existingJob != nil {
				require.NoError(t, cli.Create(context.Background(), c.existingJob))
			}

			// Test
			b.retestAndMergeOnePool(pool)

			// Delete pre-existing job
			if c.existingJob != nil {
				require.NoError(t, cli.Delete(context.Background(), c.existingJob))
			}

			ijList := &cicdv1.IntegrationJobList{}
			require.NoError(t, cli.List(context.Background(), ijList))

			if c.expectedIJRefPulls == nil {
				require.Empty(t, ijList.Items)
			} else {
				require.Len(t, ijList.Items, 1)
				require.Equal(t, c.expectedIJRefPulls, ijList.Items[0].Spec.Refs.Pulls, "IntegrationJobs")
			}

			if c.expectedBatchCreated {
				require.NotNilf(t, pool.CurrentBatch, "Current batch")
				require.NotEmpty(t, pool.CurrentBatch.Job.Name, "Current batch job")
				require.NotEmpty(t, pool.CurrentBatch.Job.Namespace, "Current batch job")

				require.Equal(t, c.prs, pool.CurrentBatch.PRs)
			} else {
				require.Nil(t, pool.CurrentBatch, "Current batch")
			}

			for _, pr := range c.prs {
				if c.expectedPRMerged {
					require.False(t, gitfake.Repos[ic.Spec.Git.Repository].PullRequests[pr.ID].Mergeable)
					require.Equal(t, git.PullRequestStateClosed, gitfake.Repos[ic.Spec.Git.Repository].PullRequests[pr.ID].State)
				} else {
					require.True(t, gitfake.Repos[ic.Spec.Git.Repository].PullRequests[pr.ID].Mergeable)
					require.Equal(t, git.PullRequestStateOpen, gitfake.Repos[ic.Spec.Git.Repository].PullRequests[pr.ID].State)
				}
			}
		})
	}
}

func TestBlocker_handleBatch(t *testing.T) {
	ic, cli := mergeTestConfig()
	gitCli, _ := utils.GetGitCli(ic, cli)
	b := New(cli)
	pool := NewPRPool(testICNamespace, testICName)

	// TEST 1 - Batch IJ fails
	t.Run("batch_ij_fails", func(t *testing.T) {
		ij1 := &cicdv1.IntegrationJob{
			ObjectMeta: metav1.ObjectMeta{Name: "test-ij-1", Namespace: testICNamespace},
		}
		_ = cli.Create(context.Background(), ij1)
		ij1.Status.State = cicdv1.IntegrationJobStateFailed
		_ = cli.Status().Update(context.Background(), ij1)

		pool.CurrentBatch = &Batch{
			PRs: []*PullRequest{{PullRequest: git.PullRequest{ID: 12}},
				{PullRequest: git.PullRequest{ID: 23}},
				{PullRequest: git.PullRequest{ID: 37}}},
			Job: types.NamespacedName{Name: "test-ij-1", Namespace: testICNamespace},
		}
		if err := b.handleBatch(pool, ic, gitCli); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 2, len(pool.CurrentBatch.PRs), "CurrentBatch gets rid of the last PR")
	})

	// TEST 2 - Batch IJ Succeeds
	t.Run("batch_ij_succeeds", func(t *testing.T) {
		ij2 := &cicdv1.IntegrationJob{
			ObjectMeta: metav1.ObjectMeta{Name: "test-ij-2", Namespace: testICNamespace},
		}
		_ = cli.Create(context.Background(), ij2)
		ij2.Status.State = cicdv1.IntegrationJobStateCompleted
		_ = cli.Status().Update(context.Background(), ij2)

		pool.CurrentBatch = &Batch{
			PRs: []*PullRequest{{PullRequest: git.PullRequest{ID: 12}}},
			Job: types.NamespacedName{Name: "test-ij-2", Namespace: testICNamespace},
		}
		if err := b.handleBatch(pool, ic, gitCli); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, pool.CurrentBatch == nil, "CurrentBatch cleared")
	})
}

func TestBlocker_mergePullRequest(t *testing.T) {
	tc := map[string]struct {
		pr             git.PullRequest
		commits        []git.Commit
		commitTemplate string

		expectedCommitMessage string
		errorOccurs           bool
		errorMessage          string
	}{
		"default": {
			pr: git.PullRequest{
				ID:    5,
				Title: "[feat] Add feature",
				Head:  git.Head{Sha: testSHA},
				Base:  git.Base{Ref: "master"},
			},
			expectedCommitMessage: "[feat] Add feature(#5)",
		},
		"commitTemplate": {
			pr: git.PullRequest{
				ID:    5,
				Title: "[feat] Add feature",
				Head:  git.Head{Sha: testSHA},
				Base:  git.Base{Ref: "master"},
			},
			commits: []git.Commit{
				{SHA: "7523ffa4cc506f4ab2a346a5c2d8eb369d0fdb30", Message: "[fix] Fix bugs", Author: git.User{Name: "author", Email: "author@tmax.co.kr"}, Committer: git.User{Name: "committer", Email: "committer@tmax.co.kr"}},
				{SHA: "3bede531bd0bbe8d3735f2642193fb33800149e0", Message: "[feat] Add features", Author: git.User{Name: "author2", Email: "author2@tmax.co.kr"}, Committer: git.User{Name: "committer2", Email: "committer2@tmax.co.kr"}},
			},
			commitTemplate: `{{.Title}}(#{{.ID}})
{{range .Commits}}
{{.SHA}}
{{.Message}}
Author: {{.Author.Name}}({{.Author.Email}})
Committer: {{.Committer.Name}}({{.Committer.Email}})
{{end}}`,
			expectedCommitMessage: `[feat] Add feature(#5)

7523ffa4cc506f4ab2a346a5c2d8eb369d0fdb30
[fix] Fix bugs
Author: author(author@tmax.co.kr)
Committer: committer(committer@tmax.co.kr)

3bede531bd0bbe8d3735f2642193fb33800149e0
[feat] Add features
Author: author2(author2@tmax.co.kr)
Committer: committer2(committer2@tmax.co.kr)
`,
		},
		"commitTemplateError": {
			pr: git.PullRequest{
				ID:    5,
				Title: "[feat] Add feature",
				Head:  git.Head{Sha: testSHA},
				Base:  git.Base{Ref: "master"},
			},
			commits: []git.Commit{
				{SHA: "7523ffa4cc506f4ab2a346a5c2d8eb369d0fdb30", Message: "[fix] Fix bugs", Author: git.User{Name: "author", Email: "author@tmax.co.kr"}, Committer: git.User{Name: "committer", Email: "committer@tmax.co.kr"}},
				{SHA: "7523ffa4cc506f4ab2a346a5c2d8eb369d0fdb30", Message: "[feat] Add features", Author: git.User{Name: "author2", Email: "author2@tmax.co.kr"}, Committer: git.User{Name: "committer2", Email: "committer2@tmax.co.kr"}},
			},
			commitTemplate: "{{.Titleeeeee}}({{.ID}})\n\nMessage",

			errorOccurs:  true,
			errorMessage: "template: :1:2: executing \"\" at <.Titleeeeee>: can't evaluate field Titleeeeee in type *blocker.PullRequest",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			ic, cli := mergeTestConfig()
			ic.Spec.MergeConfig.CommitTemplate = c.commitTemplate

			gitCli, err := utils.GetGitCli(ic, cli)
			require.NoError(t, err)
			b := New(cli)

			gitfake.Repos = map[string]*gitfake.Repo{
				ic.Spec.Git.Repository: {
					PullRequests: map[int]*git.PullRequest{
						c.pr.ID: &c.pr,
					},
					PullRequestCommits: map[int][]git.Commit{
						c.pr.ID: c.commits,
					},
					Commits: map[string][]git.Commit{},
				},
			}

			err = b.mergePullRequest(&PullRequest{PullRequest: c.pr}, ic, gitCli)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedCommitMessage, gitfake.Repos[ic.Spec.Git.Repository].Commits[c.pr.Base.Ref][0].Message)
			}
		})
	}
}

type getMergeMethodTestCase struct {
	Labels         []git.IssueLabel
	ICMethod       git.MergeMethod
	ExpectedMethod git.MergeMethod
}

func TestGetMergeMethod(t *testing.T) {
	tc := map[string]getMergeMethodTestCase{
		"followICMerge": {
			Labels:         []git.IssueLabel{},
			ICMethod:       git.MergeMethodMerge,
			ExpectedMethod: git.MergeMethodMerge,
		},
		"followICSquash": {
			Labels:         []git.IssueLabel{},
			ICMethod:       git.MergeMethodSquash,
			ExpectedMethod: git.MergeMethodSquash,
		},
		"globalMerge": {
			Labels:         []git.IssueLabel{{Name: "global/merge-merge"}},
			ICMethod:       git.MergeMethodMerge,
			ExpectedMethod: git.MergeMethodMerge,
		},
		"globalMerge2": {
			Labels:         []git.IssueLabel{{Name: "global/merge-merge"}},
			ICMethod:       git.MergeMethodSquash,
			ExpectedMethod: git.MergeMethodMerge,
		},
		"globalSquash": {
			Labels:         []git.IssueLabel{{Name: "global/merge-squash"}},
			ICMethod:       git.MergeMethodSquash,
			ExpectedMethod: git.MergeMethodSquash,
		},
		"globalSquash2": {
			Labels:         []git.IssueLabel{{Name: "global/merge-squash"}},
			ICMethod:       git.MergeMethodMerge,
			ExpectedMethod: git.MergeMethodSquash,
		},
	}

	configs.MergeKindMergeLabel = "global/merge-merge"
	configs.MergeKindSquashLabel = "global/merge-squash"

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			pr := &PullRequest{}
			pr.Labels = c.Labels
			ic := &cicdv1.IntegrationConfig{}
			ic.Spec.MergeConfig = &cicdv1.MergeConfig{Method: c.ICMethod}

			method := getMergeMethod(pr, ic)

			assert.Equal(t, string(c.ExpectedMethod), string(method))
		})
	}
}

func mergeTestConfig() (*cicdv1.IntegrationConfig, client.Client) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}

	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))
	ic := &cicdv1.IntegrationConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testICName,
			Namespace: testICNamespace,
		},
		Spec: cicdv1.IntegrationConfigSpec{
			Git: cicdv1.GitConfig{
				Type:       cicdv1.GitTypeFake,
				Repository: testRepo,
				Token:      &cicdv1.GitToken{Value: "dummy"},
			},
			MergeConfig: &cicdv1.MergeConfig{
				Query: cicdv1.MergeQuery{
					Labels:          []string{},
					ApproveRequired: true,
					Checks:          []string{"test-1"},
				},
			},
			Jobs: cicdv1.IntegrationConfigJobs{
				PreSubmit: []cicdv1.Job{
					{Container: corev1.Container{Name: "test-1"}},
				},
			},
		},
	}
	return ic, fake.NewFakeClientWithScheme(s, ic)
}
