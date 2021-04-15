package blocker

import (
	"github.com/bmizerany/assert"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"testing"
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
	result, msg := checkConditions(query, pr)

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
	result, msg = checkConditions(query, pr)

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
	result, msg = checkConditions(query, pr)

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
	result, msg = checkConditions(query, pr)

	assert.Equal(t, false, result, "Result")
	assert.Equal(t, "Label [approved] is required.", msg, "Message")
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
