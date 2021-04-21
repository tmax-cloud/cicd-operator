package blocker

import (
	"encoding/json"
	"fmt"
	"github.com/bmizerany/assert"
	"github.com/gorilla/mux"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"github.com/tmax-cloud/cicd-operator/pkg/git/github"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"net/http"
	"net/http/httptest"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
)

var testSyncStatusPR = github.PullRequest{
	Head: struct {
		Ref string `json:"ref"`
		Sha string `json:"sha"`
	}{Ref: "newnew", Sha: git.FakeSha},
	Base: struct {
		Ref string `json:"ref"`
		Sha string `json:"sha"`
	}{Ref: "master", Sha: git.FakeSha},
	Mergeable: true,
}
var testSyncStatusStatuses []github.CommitStatusResponse

func TestBlocker_syncMergePoolStatus(t *testing.T) {
	fakeCli, ic, _ := syncStatusTestEnv()
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

	// Test 1
	blocker.syncMergePoolStatus()
	assert.Equal(t, 1, len(pool.PullRequests), "PR length")
	assert.Equal(t, 0, len(pool.MergePool[git.CommitStatusStatePending]), "Pending length")
	assert.Equal(t, 0, len(pool.MergePool[git.CommitStatusStateSuccess]), "Success length")
	assert.Equal(t, "Label [approved,lgtm] is required.", pool.PullRequests[25].BlockerDescription, "Blocker status description")

	// Test 2
	testSyncStatusPR.Mergeable = false
	testSyncStatusPR.Labels = []struct {
		Name string `json:"name"`
	}{{Name: "lgtm"}, {Name: "approved"}}
	pool.MergePool[git.CommitStatusStatePending][25] = pr
	blocker.syncMergePoolStatus()
	assert.Equal(t, 1, len(pool.PullRequests), "PR length")
	assert.Equal(t, 1, len(pool.MergePool[git.CommitStatusStatePending]), "Pending length")
	assert.Equal(t, 0, len(pool.MergePool[git.CommitStatusStateSuccess]), "Success length")
	assert.Equal(t, "Merge conflicts exist. Checks [test-unit] are not successful.", pool.PullRequests[25].BlockerDescription, "Blocker status description")

	// Test 3
	testSyncStatusPR.Mergeable = true
	testSyncStatusStatuses = append(testSyncStatusStatuses, github.CommitStatusResponse{Context: "test-unit", State: "success"})
	pool.MergePool[git.CommitStatusStatePending][25] = pr
	blocker.syncMergePoolStatus()
	assert.Equal(t, 1, len(pool.PullRequests), "PR length")
	assert.Equal(t, 0, len(pool.MergePool[git.CommitStatusStatePending]), "Pending length")
	assert.Equal(t, 1, len(pool.MergePool[git.CommitStatusStateSuccess]), "Success length")
	assert.Equal(t, "In merge pool.", pool.PullRequests[25].BlockerDescription, "Blocker status description")
}

func syncStatusTestEnv() (client.Client, *cicdv1.IntegrationConfig, string) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(req.URL.String()))
	})
	r.HandleFunc("/repos/{org}/{repo}/pulls/{id}", func(w http.ResponseWriter, _ *http.Request) {
		b, _ := json.Marshal(testSyncStatusPR)
		_, _ = w.Write(b)
	})
	r.HandleFunc("/repos/{org}/{repos}/commits/{sha}/statuses", func(w http.ResponseWriter, req *http.Request) {
		b, _ := json.Marshal(testSyncStatusStatuses)
		_, _ = w.Write(b)
	})
	r.HandleFunc("/repos/{org}/{repos}/statuses/{sha}", func(_ http.ResponseWriter, _ *http.Request) {})
	testSrv := httptest.NewServer(r)

	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))

	ic := &cicdv1.IntegrationConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: cicdv1.IntegrationConfigSpec{
			Git: cicdv1.GitConfig{
				Type:       "github",
				Repository: "tmax-cloud/cicd-operator",
				APIUrl:     testSrv.URL,
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

	return fake.NewFakeClientWithScheme(s, ic), ic, fmt.Sprintf("%s/%s", testSrv.URL, ic.Spec.Git.Repository)
}
