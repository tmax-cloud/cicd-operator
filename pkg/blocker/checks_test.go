package blocker

import (
	"fmt"
	"github.com/bmizerany/assert"
	"github.com/gorilla/mux"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"net/http"
	"net/http/httptest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"strings"
	"testing"
)

var (
	sampleStatusesList = "[{\"id\":1111111111,\"state\":\"pending\",\"context\":\"test-1\",\"created_at\":\"2021-04-12T08:37:32Z\",\"updated_at\":\"2021-04-12T08:37:32Z\",\"creator\":{\"login\":\"sunghyunkim3\",\"id\":1111111,\"type\":\"User\",\"site_admin\":false}}]"
	samplePR           = "{\"id\":611161419,\"html_url\":\"https://github.com/vingsu/cicd-test/pull/25\",\"number\":25,\"state\":\"open\",\"title\":\"newnew\",\"user\":{\"login\":\"cqbqdd11519\",\"id\":6166781},\"body\":\"\",\"labels\":[{\"id\":2905890093,\"node_id\":\"MDU6TGFiZWwyOTA1ODkwMDkz\",\"url\":\"https://api.github.com/repos/vingsu/cicd-test/labels/kind/test\",\"name\":\"kind/test\",\"color\":\"CF61D3\",\"default\":false,\"description\":\"\"}],\"milestone\":null,\"draft\":false,\"head\":{\"label\":\"vingsu:newnew\",\"ref\":\"newnew\",\"sha\":\"3196ccc37bcae94852079b04fcbfaf928341d6e9\",\"user\":{\"login\":\"vingsu\",\"id\":71878727},\"repo\":{\"id\":319253224,\"node_id\":\"MDEwOlJlcG9zaXRvcnkzMTkyNTMyMjQ=\",\"name\":\"cicd-test\",\"full_name\":\"vingsu/cicd-test\",\"private\":true,\"owner\":{\"login\":\"vingsu\",\"id\":71878727},\"html_url\":\"https://github.com/vingsu/cicd-test\"}},\"base\":{\"label\":\"vingsu:master\",\"ref\":\"master\",\"sha\":\"22ccae53032027186ba739dfaa473ee61a82b298\",\"user\":{\"login\":\"vingsu\",\"id\":71878727},\"repo\":{\"id\":319253224,\"node_id\":\"MDEwOlJlcG9zaXRvcnkzMTkyNTMyMjQ=\",\"name\":\"cicd-test\",\"full_name\":\"vingsu/cicd-test\",\"private\":true,\"owner\":{\"login\":\"vingsu\",\"id\":71878727},\"html_url\":\"https://github.com/vingsu/cicd-test\"}},\"merged\":false,\"mergeable\":false}"
)

var serverURL string

func TestCheckConditions(t *testing.T) {
	// Test 2
	pr := &git.PullRequest{
		Sender:    git.User{Name: "cqbqdd11519"},
		Base:      git.Base{Ref: "refs/heads/newnew"},
		Labels:    []git.IssueLabel{{Name: "lgtm"}},
		Mergeable: true,
	}
	query := cicdv1.MergeQuery{}
	result, msg := checkConditionsSimple(query, pr)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")

	// Test 3
	pr = &git.PullRequest{
		Sender:    git.User{Name: "cqbqdd11519"},
		Base:      git.Base{Ref: "refs/heads/newnew"},
		Labels:    []git.IssueLabel{{Name: "lgtm"}},
		Mergeable: true,
	}
	query = cicdv1.MergeQuery{
		Branches: []string{"master"},
	}
	result, msg = checkConditionsSimple(query, pr)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Branch [newnew] is not in branches query.", msg, "Message")

	// Test 4
	pr = &git.PullRequest{
		Sender:    git.User{Name: "cqbqdd11519"},
		Base:      git.Base{Ref: "refs/heads/newnew"},
		Labels:    []git.IssueLabel{{Name: "lgtm"}},
		Mergeable: true,
	}
	query = cicdv1.MergeQuery{
		Branches: []string{"master", "newnew"},
	}
	result, msg = checkConditionsSimple(query, pr)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")

	// Test 5
	pr = &git.PullRequest{
		Sender:    git.User{Name: "cqbqdd11519"},
		Base:      git.Base{Ref: "refs/heads/newnew"},
		Labels:    []git.IssueLabel{{Name: "lgtm"}},
		Mergeable: true,
	}
	query = cicdv1.MergeQuery{
		Branches:        []string{"master", "newnew"},
		Labels:          []string{"lgtm"},
		ApproveRequired: true,
	}
	result, msg = checkConditionsSimple(query, pr)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Label [approved] is required.", msg, "Message")
}

func TestCheckConditionsFull(t *testing.T) {
	fakeCli, ic := checkTestEnv()
	gitCli, err := utils.GetGitCli(ic, fakeCli)
	if err != nil {
		t.Fatal(err)
	}

	// Test 1
	status, msg, err := checkConditionsFull(ic.Spec.MergeConfig.Query, 25, gitCli)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, false, status, "Full status")
	assert.Equal(t, "Label [approved] is required. Merge conflicts exist. Checks [test-1] are not successful.", msg, "Full message")

	// Test 2
	samplePR = strings.ReplaceAll(samplePR, "\"mergeable\":false", "\"mergeable\":true")
	samplePR = strings.ReplaceAll(samplePR, "kind/test", "approved")
	sampleStatusesList = strings.ReplaceAll(sampleStatusesList, "pending", "success")
	status, msg, err = checkConditionsFull(ic.Spec.MergeConfig.Query, 25, gitCli)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, status, "Full status")
	assert.Equal(t, "", msg, "Full message")
}

func TestCheckBranch(t *testing.T) {
	// Test 1
	branch := "master"
	query := cicdv1.MergeQuery{
		Branches: []string{"master", "master2"},
	}
	result, msg := checkBranch(branch, query)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")

	// Test 2
	branch = "masters"
	query = cicdv1.MergeQuery{
		Branches: []string{"master", "master2"},
	}
	result, msg = checkBranch(branch, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Branch [masters] is not in branches query.", msg, "Message")

	// Test 3
	branch = "master"
	query = cicdv1.MergeQuery{
		SkipBranches: []string{"master", "master2"},
	}
	result, msg = checkBranch(branch, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Branch [master] is in skipBranches query.", msg, "Message")

	// Test 4
	branch = "masters"
	query = cicdv1.MergeQuery{
		SkipBranches: []string{"master", "master2"},
	}
	result, msg = checkBranch(branch, query)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")
}

func TestCheckAuthor(t *testing.T) {
	// Test 1
	author := "cqbqdd11519"
	query := cicdv1.MergeQuery{
		Authors: []string{"cqbqdd11519"},
	}
	result, msg := checkAuthor(author, query)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")

	// Test 2
	author = "sunghyunkim3"
	query = cicdv1.MergeQuery{
		Authors: []string{"cqbqdd11519"},
	}
	result, msg = checkAuthor(author, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Author [sunghyunkim3] is not in authors query.", msg, "Message")

	// Test 3
	author = "cqbqdd11519"
	query = cicdv1.MergeQuery{
		SkipAuthors: []string{"cqbqdd11519"},
	}
	result, msg = checkAuthor(author, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Author [cqbqdd11519] is in skipAuthors query.", msg, "Message")

	// Test 4
	author = "sunghyunkim3"
	query = cicdv1.MergeQuery{
		SkipAuthors: []string{"cqbqdd11519"},
	}
	result, msg = checkAuthor(author, query)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")

	// Test 5
	author = "sunghyunkim3"
	query = cicdv1.MergeQuery{}
	result, msg = checkAuthor(author, query)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")
}

func TestCheckLabels(t *testing.T) {
	// Test 1
	labels := []string{
		"lgtm",
		"hold",
	}
	query := cicdv1.MergeQuery{
		Labels:      []string{},
		BlockLabels: []string{},
	}
	result, msg := checkLabels(labels, query)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")

	// Test 2
	labels = []string{}
	query = cicdv1.MergeQuery{
		Labels: []string{
			"lgtm",
		},
		BlockLabels: []string{
			"hold",
		},
	}
	result, msg = checkLabels(labels, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Label [lgtm] is required.", msg, "Message")

	// Test 3
	labels = []string{
		"lgtm",
	}
	query = cicdv1.MergeQuery{
		Labels: []string{
			"lgtm",
		},
		BlockLabels: []string{
			"hold",
		},
	}
	result, msg = checkLabels(labels, query)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")

	// Test 4
	labels = []string{
		"lgtm",
		"hold",
	}
	query = cicdv1.MergeQuery{
		Labels: []string{
			"lgtm",
		},
		BlockLabels: []string{
			"hold",
		},
	}
	result, msg = checkLabels(labels, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Label [hold] is blocking the merge.", msg, "Message")

	// Test 5
	labels = []string{
		"lgtm",
		"kind/bug",
	}
	query = cicdv1.MergeQuery{
		Labels: []string{
			"lgtm",
			"approved",
		},
		BlockLabels: []string{
			"hold",
		},
	}
	result, msg = checkLabels(labels, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Label [approved] is required.", msg, "Message")

	// Test 6
	labels = []string{
		"hold",
	}
	query = cicdv1.MergeQuery{
		Labels: []string{
			"lgtm",
			"approved",
		},
		BlockLabels: []string{
			"hold",
		},
	}
	result, msg = checkLabels(labels, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Label [hold] is blocking the merge.", msg, "Message")

	// Test 7
	labels = []string{
		"lgtm",
		"approved",
		"hold",
	}
	query = cicdv1.MergeQuery{
		Labels: []string{
			"lgtm",
			"approved",
		},
		BlockLabels: []string{
			"hold",
		},
	}
	result, msg = checkLabels(labels, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Label [hold] is blocking the merge.", msg, "Message")

	// Test 8
	labels = []string{
		"lgtm",
		"approved",
	}
	query = cicdv1.MergeQuery{
		Labels: []string{
			"lgtm",
			"approved",
		},
		BlockLabels: []string{
			"hold",
		},
	}
	result, msg = checkLabels(labels, query)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")
}

func TestCheckChecks(t *testing.T) {
	// Test 1
	statuses := []git.CommitStatus{
		{Context: "test-lint", State: "pending"},
		{Context: "test-unit", State: "success"},
	}
	query := cicdv1.MergeQuery{
		Checks: []string{},
	}
	result, msg := checkChecks(statuses, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Checks [test-lint] are not successful.", msg, "Message")

	// Test 2
	statuses = []git.CommitStatus{
		{Context: "test-lint", State: "pending"},
		{Context: "test-unit", State: "success"},
	}
	query = cicdv1.MergeQuery{
		OptionalChecks: []string{
			"test-lint",
		},
	}
	result, msg = checkChecks(statuses, query)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")

	// Test 3
	statuses = []git.CommitStatus{
		{Context: "test-lint", State: "pending"},
		{Context: "test-unit", State: "success"},
	}
	query = cicdv1.MergeQuery{
		Checks: []string{
			"test-unit",
		},
	}
	result, msg = checkChecks(statuses, query)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")

	// Test 4
	statuses = []git.CommitStatus{
		{Context: "test-lint", State: "pending"},
		{Context: "test-unit", State: "pending"},
	}
	query = cicdv1.MergeQuery{
		Checks: []string{
			"test-unit",
		},
	}
	result, msg = checkChecks(statuses, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Checks [test-unit] are not successful.", msg, "Message")

	// Test 5
	statuses = []git.CommitStatus{
		{Context: "test-lint", State: "pending"},
		{Context: "test-unit", State: "pending"},
	}
	query = cicdv1.MergeQuery{}
	result, msg = checkChecks(statuses, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Checks [test-lint,test-unit] are not successful.", msg, "Message")

	// Test 6
	statuses = []git.CommitStatus{
		{Context: "test-lint", State: "pending"},
		{Context: "test-unit", State: "pending"},
	}
	query = cicdv1.MergeQuery{
		Checks: []string{
			"test-lint",
			"test-unit",
		},
	}
	result, msg = checkChecks(statuses, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Checks [test-lint,test-unit] are not successful.", msg, "Message")

	// Test 7
	statuses = []git.CommitStatus{
		{Context: blockerContext, State: "pending"},
		{Context: "test-lint", State: "pending"},
		{Context: "test-unit", State: "success"},
	}
	query = cicdv1.MergeQuery{
		OptionalChecks: []string{
			"test-lint",
		},
	}
	result, msg = checkChecks(statuses, query)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")
}

func checkTestEnv() (client.Client, *cicdv1.IntegrationConfig) {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(req.URL.String()))
	})
	r.HandleFunc("/repos/{org}/{repo}/commits/{sha}/statuses", func(w http.ResponseWriter, req *http.Request) {
		page := req.URL.Query().Get("page")
		if page == "" || page == "1" {
			w.Header().Set("Link", fmt.Sprintf("<%s/%s?state=all&per_page=100&page=2>; rel=\"next\", <%s/%s?state=all&per_page=100&page=3>; rel=\"last\"", serverURL, req.URL.Path, serverURL, req.URL.Path))
		}
		_, _ = w.Write([]byte(sampleStatusesList))
	})
	r.HandleFunc("/repos/{org}/{repo}/pulls/{id}", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(samplePR))
	})
	testSrv := httptest.NewServer(r)
	serverURL = testSrv.URL

	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))

	ic := &cicdv1.IntegrationConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ic",
			Namespace: "default",
		},
		Spec: cicdv1.IntegrationConfigSpec{
			Git: cicdv1.GitConfig{
				Type:       "github",
				Repository: "tmax-cloud/cicd-test",
				APIUrl:     serverURL,
				Token:      &cicdv1.GitToken{Value: "dummy"},
			},
			MergeConfig: &cicdv1.MergeConfig{
				Query: cicdv1.MergeQuery{
					Labels:          []string{},
					ApproveRequired: true,
					Checks:          []string{"test-1"},
				},
			},
		},
	}

	return fake.NewFakeClientWithScheme(s, ic), ic
}
