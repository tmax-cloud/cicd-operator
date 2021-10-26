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
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	gitfake "github.com/tmax-cloud/cicd-operator/pkg/git/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestBlocker_syncPRs(t *testing.T) {
	fakeCli, ic := syncPoolTestEnv()
	blocker := New(fakeCli)
	pools := blocker.Pools

	// Initial
	blocker.syncPRs()
	assert.Equal(t, 1, len(pools[genPoolKey(ic)].PullRequests), "PRList length")
	assert.Equal(t, 0, len(pools[genPoolKey(ic)].MergePool[git.CommitStatusStatePending]), "Pending merge pool length")

	// Hold
	gitfake.Repos[testRepo].PullRequests[testPRID].Labels = append(gitfake.Repos[testRepo].PullRequests[testPRID].Labels, git.IssueLabel{Name: "hold"})
	blocker.syncPRs()
	assert.Equal(t, 1, len(pools[genPoolKey(ic)].PullRequests), "PRList length")
	assert.Equal(t, 0, len(pools[genPoolKey(ic)].MergePool[git.CommitStatusStatePending]), "Pending merge pool length")

	// LGTM
	if err := gitfake.DeleteLabel(testRepo, testPRID, "hold"); err != nil {
		t.Fatal(err)
	}

	gitfake.Repos[testRepo].PullRequests[testPRID].Labels = append(gitfake.Repos[testRepo].PullRequests[testPRID].Labels, git.IssueLabel{Name: "lgtm"})
	blocker.syncPRs()
	assert.Equal(t, 1, len(pools[genPoolKey(ic)].PullRequests), "PRList length")
	assert.Equal(t, 1, len(pools[genPoolKey(ic)].MergePool[git.CommitStatusStatePending]), "Pending merge pool length")

	// New PR
	newPRID := 24
	gitfake.Repos[testRepo].PullRequests[newPRID] = &git.PullRequest{
		ID:     newPRID,
		Title:  "[fix] Fix bugs",
		State:  "open",
		Author: git.User{Name: "test"},
		Base:   git.Base{Ref: "master"},
		Labels: []git.IssueLabel{
			{Name: "kind/bug"},
			{Name: "size/L"},
		},
	}
	blocker.syncPRs()
	assert.Equal(t, 2, len(pools[genPoolKey(ic)].PullRequests), "PRList length")
	assert.Equal(t, 1, len(pools[genPoolKey(ic)].MergePool[git.CommitStatusStatePending]), "Pending merge pool length")

	// LGTM 2
	gitfake.Repos[testRepo].PullRequests[newPRID].Labels = []git.IssueLabel{{Name: "lgtm"}}
	blocker.syncPRs()
	assert.Equal(t, 2, len(pools[genPoolKey(ic)].PullRequests), "PRList length")
	assert.Equal(t, 2, len(pools[genPoolKey(ic)].MergePool[git.CommitStatusStatePending]), "Pending merge pool length")

	// Deleted PR (closed maybe)
	delete(gitfake.Repos[testRepo].PullRequests, testPRID)
	delete(gitfake.Repos[testRepo].PullRequests, newPRID)
	blocker.syncPRs()
	assert.Equal(t, 0, len(pools[genPoolKey(ic)].PullRequests), "PRList length")
	assert.Equal(t, 0, len(pools[genPoolKey(ic)].MergePool[git.CommitStatusStatePending]), "Pending merge pool length")

	// Deleted ic
	if err := fakeCli.Delete(context.Background(), &cicdv1.IntegrationConfig{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"}}); err != nil {
		t.Fatal(err)
	}
	blocker.syncPRs()
	assert.Equal(t, 0, len(pools), "IC length")
}

func syncPoolTestEnv() (client.Client, *cicdv1.IntegrationConfig) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}

	gitfake.Repos = map[string]*gitfake.Repo{
		testRepo: {
			Webhooks:     map[int]*git.WebhookEntry{},
			UserCanWrite: map[string]bool{},

			PullRequests:   map[int]*git.PullRequest{},
			CommitStatuses: map[string][]git.CommitStatus{},
			Comments:       map[int][]git.IssueComment{},
		},
	}

	gitfake.Repos[testRepo].PullRequests[testPRID] = &git.PullRequest{
		ID:     testPRID,
		Title:  "[feat] New feature",
		State:  "open",
		Author: git.User{Name: "test"},
		Base:   git.Base{Ref: "master"},
		Labels: []git.IssueLabel{
			{Name: "kind/feat"},
			{Name: "size/L"},
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
					Branches:       []string{"master"},
					Labels:         []string{"lgtm"},
					BlockLabels:    []string{"hold"},
					OptionalChecks: []string{"send-email"},
				},
			},
			Jobs: cicdv1.IntegrationConfigJobs{},
		},
	}

	return fake.NewClientBuilder().WithScheme(s).WithObjects(ic).Build(), ic
}
