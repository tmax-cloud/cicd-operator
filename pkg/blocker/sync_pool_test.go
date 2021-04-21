package blocker

import (
	"context"
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

var testSyncPoolPRs []github.PullRequest

func TestBlocker_syncPRs(t *testing.T) {
	fakeCli, repoURL := syncPoolTestEnv()
	blocker := New(fakeCli)
	pools := blocker.Pools

	// Initial
	blocker.syncPRs()
	assert.Equal(t, 1, len(pools[poolKey(repoURL)].PullRequests), "PRList length")
	assert.Equal(t, 0, len(pools[poolKey(repoURL)].MergePool[git.CommitStatusStatePending]), "Unknown merge pool length")

	// Hold
	testSyncPoolPRs[0].Labels = append(testSyncPoolPRs[0].Labels, struct {
		Name string `json:"name"`
	}{Name: "hold"})
	blocker.syncPRs()
	assert.Equal(t, 1, len(pools[poolKey(repoURL)].PullRequests), "PRList length")
	assert.Equal(t, 0, len(pools[poolKey(repoURL)].MergePool[git.CommitStatusStatePending]), "Unknown merge pool length")

	// LGTM
	testSyncPoolPRs[0].Labels = []struct {
		Name string `json:"name"`
	}{{Name: "lgtm"}}
	blocker.syncPRs()
	assert.Equal(t, 1, len(pools[poolKey(repoURL)].PullRequests), "PRList length")
	assert.Equal(t, 1, len(pools[poolKey(repoURL)].MergePool[git.CommitStatusStatePending]), "Unknown merge pool length")

	// New PR
	testSyncPoolPRs = append(testSyncPoolPRs, github.PullRequest{
		Number: 24,
		Title:  "[fix] Fix bugs",
		State:  "open",
		User:   github.User{Name: "test"},
		Base: struct {
			Ref string `json:"ref"`
			Sha string `json:"sha"`
		}{Ref: "master"},
		Labels: []struct {
			Name string `json:"name"`
		}{
			{Name: "kind/bug"},
			{Name: "size/L"},
		},
	})
	blocker.syncPRs()
	assert.Equal(t, 2, len(pools[poolKey(repoURL)].PullRequests), "PRList length")
	assert.Equal(t, 1, len(pools[poolKey(repoURL)].MergePool[git.CommitStatusStatePending]), "Unknown merge pool length")

	// LGTM 2
	testSyncPoolPRs[1].Labels = []struct {
		Name string `json:"name"`
	}{{Name: "lgtm"}}
	blocker.syncPRs()
	assert.Equal(t, 2, len(pools[poolKey(repoURL)].PullRequests), "PRList length")
	assert.Equal(t, 2, len(pools[poolKey(repoURL)].MergePool[git.CommitStatusStatePending]), "Unknown merge pool length")

	// Deleted PR (closed maybe)
	testSyncPoolPRs = nil
	blocker.syncPRs()
	assert.Equal(t, 0, len(pools[poolKey(repoURL)].PullRequests), "PRList length")
	assert.Equal(t, 0, len(pools[poolKey(repoURL)].MergePool[git.CommitStatusStatePending]), "Unknown merge pool length")

	// Deleted ic
	if err := fakeCli.Delete(context.Background(), &cicdv1.IntegrationConfig{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"}}); err != nil {
		t.Fatal(err)
	}
	blocker.syncPRs()
	assert.Equal(t, 0, len(pools), "IC length")
}

func syncPoolTestEnv() (client.Client, string) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(req.URL.String()))
	})
	r.HandleFunc("/repos/{org}/{repo}/pulls", func(w http.ResponseWriter, _ *http.Request) {
		b, _ := json.Marshal(testSyncPoolPRs)
		_, _ = w.Write(b)
	})
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
					Branches:       []string{"master"},
					Labels:         []string{"lgtm"},
					BlockLabels:    []string{"hold"},
					OptionalChecks: []string{"send-email"},
				},
			},
			Jobs: cicdv1.IntegrationConfigJobs{},
		},
	}

	testSyncPoolPRs = []github.PullRequest{{
		Number: 23,
		Title:  "[feat] New feature",
		State:  "open",
		User:   github.User{Name: "test"},
		Draft:  false,
		Base: struct {
			Ref string `json:"ref"`
			Sha string `json:"sha"`
		}{Ref: "master"},
		Labels: []struct {
			Name string `json:"name"`
		}{
			{Name: "kind/feat"},
			{Name: "size/L"},
		},
	}}

	return fake.NewFakeClientWithScheme(s, ic), fmt.Sprintf("%s/%s", testSrv.URL, ic.Spec.Git.Repository)
}
