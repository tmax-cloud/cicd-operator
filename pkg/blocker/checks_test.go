package blocker

import (
	"github.com/bmizerany/assert"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	gitfake "github.com/tmax-cloud/cicd-operator/pkg/git/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

const (
	testRepo = "tmax-cloud/cicd-test"
	testPRID = 25
	testSHA  = "1896d4e0deaed7cda867f42935934ee13e370012"
)

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
	status, removeFromMergePool, msg, err := checkConditionsFull(ic.Spec.MergeConfig.Query, testPRID, gitCli)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, false, status, "Full status")
	assert.Equal(t, true, removeFromMergePool, "Remove from merge pool")
	assert.Equal(t, "Label [approved] is required.", msg, "Full message")

	// Test 2
	gitfake.Repos[testRepo].PullRequests[testPRID].Labels = []git.IssueLabel{{Name: "approved"}}
	status, removeFromMergePool, msg, err = checkConditionsFull(ic.Spec.MergeConfig.Query, testPRID, gitCli)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, false, status, "Full status")
	assert.Equal(t, false, removeFromMergePool, "Remove from merge pool")
	assert.Equal(t, "Merge conflicts exist. Checks [test-1] are not successful.", msg, "Full message")

	// Test 3
	gitfake.Repos[testRepo].PullRequests[testPRID].Mergeable = true
	gitfake.Repos[testRepo].CommitStatuses[testSHA] = []git.CommitStatus{{Context: "test-1", State: "success"}}
	status, removeFromMergePool, msg, err = checkConditionsFull(ic.Spec.MergeConfig.Query, testPRID, gitCli)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, status, "Full status")
	assert.Equal(t, false, removeFromMergePool, "Remove from merge pool")
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
	labels := map[string]struct{}{
		"lgtm": {},
		"hold": {},
	}
	query := cicdv1.MergeQuery{
		Labels:      []string{},
		BlockLabels: []string{},
	}
	result, msg := checkLabels(labels, query)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")

	// Test 2
	labels = nil
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
	labels = map[string]struct{}{
		"lgtm": {},
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
	labels = map[string]struct{}{
		"lgtm": {},
		"hold": {},
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
	labels = map[string]struct{}{
		"lgtm":     {},
		"kind/bug": {},
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
	labels = map[string]struct{}{
		"hold": {},
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
	assert.Equal(t, "Label [approved,lgtm] is required. Label [hold] is blocking the merge.", msg, "Message")

	// Test 7
	labels = map[string]struct{}{
		"lgtm":     {},
		"approved": {},
		"hold":     {},
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
	labels = map[string]struct{}{
		"lgtm":     {},
		"approved": {},
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
	statuses := map[string]git.CommitStatusState{
		"test-lint": "pending",
		"test-unit": "success",
	}
	query := cicdv1.MergeQuery{
		Checks: []string{},
	}
	result, msg := checkChecks(statuses, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Checks [test-lint] are not successful.", msg, "Message")

	// Test 2
	statuses = map[string]git.CommitStatusState{
		"test-lint": "pending",
		"test-unit": "success",
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
	statuses = map[string]git.CommitStatusState{
		"test-lint": "pending",
		"test-unit": "success",
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
	statuses = map[string]git.CommitStatusState{
		"test-lint": "pending",
		"test-unit": "pending",
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
	statuses = map[string]git.CommitStatusState{
		"test-lint": "pending",
		"test-unit": "pending",
	}
	query = cicdv1.MergeQuery{}
	result, msg = checkChecks(statuses, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Checks [test-lint,test-unit] are not successful.", msg, "Message")

	// Test 6
	statuses = map[string]git.CommitStatusState{
		"test-lint": "pending",
		"test-unit": "pending",
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
	statuses = map[string]git.CommitStatusState{
		blockerContext: "pending",
		"test-lint":    "pending",
		"test-unit":    "success",
	}
	query = cicdv1.MergeQuery{
		OptionalChecks: []string{
			"test-lint",
		},
	}
	result, msg = checkChecks(statuses, query)

	assert.Equal(t, true, result, "Result")
	assert.Equal(t, "", msg, "Message")

	// Test 8
	statuses = map[string]git.CommitStatusState{
		blockerContext: "pending",
		"test-lint":    "pending",
	}
	query = cicdv1.MergeQuery{
		Checks: []string{
			"test-unit",
		},
	}
	result, msg = checkChecks(statuses, query)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Checks [test-unit] are not successful.", msg, "Message")
}

func checkTestEnv() (client.Client, *cicdv1.IntegrationConfig) {
	gitfake.Repos = map[string]gitfake.Repo{
		testRepo: {
			Webhooks:     map[int]git.WebhookEntry{},
			UserCanWrite: map[string]bool{},

			PullRequests:   map[int]*git.PullRequest{},
			CommitStatuses: map[string][]git.CommitStatus{},
			Comments:       map[int][]git.IssueComment{},
		},
	}
	gitfake.Repos[testRepo].PullRequests[testPRID] = &git.PullRequest{
		Head:   git.Head{Sha: testSHA},
		Labels: []git.IssueLabel{{Name: "kind/bug"}},
	}
	gitfake.Repos[testRepo].CommitStatuses[testSHA] = nil

	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))

	ic := &cicdv1.IntegrationConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ic",
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
					Labels:          []string{},
					ApproveRequired: true,
					Checks:          []string{"test-1"},
				},
			},
		},
	}

	return fake.NewFakeClientWithScheme(s, ic), ic
}
