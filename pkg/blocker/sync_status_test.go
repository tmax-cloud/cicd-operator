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
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	gitfake "github.com/tmax-cloud/cicd-operator/pkg/git/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
)

func TestBlocker_syncMergePoolStatus(t *testing.T) {
	fakeCli, ic := syncStatusTestEnv()
	blocker := New(fakeCli)

	key := genPoolKey(ic)
	blocker.Pools[key] = NewPRPool(ic.Namespace, ic.Name)

	pool := blocker.Pools[key]
	pool.MergePool = NewMergePool()

	pr := &PullRequest{
		PullRequest: git.PullRequest{
			ID: 25,
			Head: struct {
				Ref string
				Sha string
			}{Ref: "newnew", Sha: git.FakeSha},
			Base: struct {
				Ref string
				Sha string
			}{Ref: "master", Sha: git.FakeSha},
		},
		BlockerStatus: git.CommitStatusStatePending,
	}
	pool.PullRequests[25] = pr
	pool.MergePool[git.CommitStatusStatePending][25] = pr
	gitfake.Repos[testRepo].CommitStatuses[testSHA] = []git.CommitStatus{{Context: ""}}

	// Test 1
	blocker.syncMergePoolStatus()
	assert.Equal(t, 1, len(pool.PullRequests), "PR length")
	assert.Equal(t, 0, len(pool.MergePool[git.CommitStatusStatePending]), "Pending length")
	assert.Equal(t, 0, len(pool.MergePool[git.CommitStatusStateSuccess]), "Success length")
	assert.Equal(t, "Label [approved,lgtm] is required.", pool.PullRequests[25].BlockerDescription, "Blocker status description")

	// Test 2
	gitfake.Repos[testRepo].PullRequests[testPRID].Mergeable = false
	gitfake.Repos[testRepo].PullRequests[testPRID].Labels = []git.IssueLabel{
		{Name: "lgtm"},
		{Name: "approved"},
	}
	pool.MergePool[git.CommitStatusStatePending][testPRID] = pr
	blocker.syncMergePoolStatus()
	assert.Equal(t, 1, len(pool.PullRequests), "PR length")
	assert.Equal(t, 1, len(pool.MergePool[git.CommitStatusStatePending]), "Pending length")
	assert.Equal(t, 0, len(pool.MergePool[git.CommitStatusStateSuccess]), "Success length")
	assert.Equal(t, "Merge conflicts exist. Checks [test-unit] are not successful.", pool.PullRequests[25].BlockerDescription, "Blocker status description")

	// Test 3
	gitfake.Repos[testRepo].PullRequests[testPRID].Mergeable = true
	gitfake.Repos[testRepo].CommitStatuses[testSHA] = []git.CommitStatus{{Context: "test-unit", State: "success"}}
	pool.MergePool[git.CommitStatusStatePending][testPRID] = pr
	blocker.syncMergePoolStatus()
	assert.Equal(t, 1, len(pool.PullRequests), "PR length")
	assert.Equal(t, 0, len(pool.MergePool[git.CommitStatusStatePending]), "Pending length")
	assert.Equal(t, 1, len(pool.MergePool[git.CommitStatusStateSuccess]), "Success length")
	assert.Equal(t, "In merge pool.", pool.PullRequests[25].BlockerDescription, "Blocker status description")
}

func syncStatusTestEnv() (client.Client, *cicdv1.IntegrationConfig) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}

	gitfake.Repos = map[string]*gitfake.Repo{
		testRepo: {
			Webhooks:     map[int]*git.WebhookEntry{},
			UserCanWrite: map[string]bool{},

			PullRequests: map[int]*git.PullRequest{
				testPRID: {
					ID:        testPRID,
					Head:      git.Head{Ref: "newnew", Sha: testSHA},
					Base:      git.Base{Ref: "master", Sha: git.FakeSha},
					Mergeable: true,
				},
			},
			CommitStatuses: map[string][]git.CommitStatus{},
			Comments:       map[int][]git.IssueComment{},
		},
	}

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
				Token:      &cicdv1.GitToken{Value: "dummy"},
			},
			MergeConfig: &cicdv1.MergeConfig{
				Query: cicdv1.MergeQuery{
					Branches:        []string{"master"},
					Labels:          []string{"lgtm"},
					BlockLabels:     []string{"hold"},
					Checks:          []string{"test-unit"},
					ApproveRequired: true,
				},
			},
			Jobs: cicdv1.IntegrationConfigJobs{},
		},
	}

	return fake.NewFakeClientWithScheme(s, ic), ic
}
