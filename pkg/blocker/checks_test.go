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
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

const (
	testRepo = "tmax-cloud/cicd-test"
	testPRID = 25
	testSHA  = "1896d4e0deaed7cda867f42935934ee13e370012"
)

type checkConditionTestCase struct {
	PR    *git.PullRequest
	Query cicdv1.MergeQuery

	ExpectedResult  bool
	ExpectedMessage string
}

func TestCheckConditions(t *testing.T) {
	tc := map[string]checkConditionTestCase{
		"success": {
			PR: &git.PullRequest{
				Author:    git.User{Name: "cqbqdd11519"},
				Base:      git.Base{Ref: "refs/heads/newnew"},
				Labels:    []git.IssueLabel{{Name: "lgtm"}},
				Mergeable: true,
			},
			Query:           cicdv1.MergeQuery{},
			ExpectedResult:  true,
			ExpectedMessage: "",
		},
		"failBranch": {
			PR: &git.PullRequest{
				Author:    git.User{Name: "cqbqdd11519"},
				Base:      git.Base{Ref: "refs/heads/newnew"},
				Labels:    []git.IssueLabel{{Name: "lgtm"}},
				Mergeable: true,
			},
			Query: cicdv1.MergeQuery{
				Branches: []string{"master"},
			},
			ExpectedResult:  false,
			ExpectedMessage: "Branch [newnew] is not in branches query.",
		},
		"successBranch": {
			PR: &git.PullRequest{
				Author:    git.User{Name: "cqbqdd11519"},
				Base:      git.Base{Ref: "refs/heads/newnew"},
				Labels:    []git.IssueLabel{{Name: "lgtm"}},
				Mergeable: true,
			},
			Query: cicdv1.MergeQuery{
				Branches: []string{"master", "newnew"},
			},
			ExpectedResult:  true,
			ExpectedMessage: "",
		},
		"failLabel": {
			PR: &git.PullRequest{
				Author:    git.User{Name: "cqbqdd11519"},
				Base:      git.Base{Ref: "refs/heads/newnew"},
				Labels:    []git.IssueLabel{{Name: "lgtm"}},
				Mergeable: true,
			},
			Query: cicdv1.MergeQuery{
				Branches:        []string{"master", "newnew"},
				Labels:          []string{"lgtm"},
				ApproveRequired: true,
			},
			ExpectedResult:  false,
			ExpectedMessage: "Label [approved] is required.",
		},
		"failGlobalBlock": {
			PR: &git.PullRequest{
				Author:    git.User{Name: "cqbqdd11519"},
				Base:      git.Base{Ref: "refs/heads/newnew"},
				Labels:    []git.IssueLabel{{Name: "lgtm"}, {Name: "global/block-label"}},
				Mergeable: true,
			},
			Query: cicdv1.MergeQuery{
				Branches:        []string{"master", "newnew"},
				Labels:          []string{"lgtm"},
				ApproveRequired: false,
			},
			ExpectedResult:  false,
			ExpectedMessage: "Label [global/block-label] is blocking the merge.",
		},
	}

	// For test 'failGlobalBlock'
	configs.MergeBlockLabel = "global/block-label"

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			result, msg := checkConditionsSimple(c.Query, c.PR)
			assert.Equal(t, c.ExpectedResult, result)
			assert.Equal(t, c.ExpectedMessage, msg)
		})
	}
}

type checkConditionsFullTestCase struct {
	FuncPre func(*PullRequest)

	ExpectedResult         bool
	ExpectedRemoveFromPool bool
	ExpectedMessage        string
}

func TestCheckConditionsFull(t *testing.T) {
	tc := map[string]checkConditionsFullTestCase{
		"failLabel": {
			FuncPre:                func(pr *PullRequest) {},
			ExpectedResult:         false,
			ExpectedRemoveFromPool: true,
			ExpectedMessage:        "Label [approved] is required.",
		},
		"failCheck": {
			FuncPre: func(pr *PullRequest) {
				pr.Labels = []git.IssueLabel{{Name: "approved"}}
			},
			ExpectedResult:         false,
			ExpectedRemoveFromPool: false,
			ExpectedMessage:        "Merge conflicts exist. Checks [test-1] are not successful.",
		},
		"success": {
			FuncPre: func(pr *PullRequest) {
				pr.Mergeable = true
				pr.Labels = []git.IssueLabel{{Name: "approved"}}
				pr.Statuses["test-1"] = git.CommitStatus{State: "success"}
			},
			ExpectedResult:         true,
			ExpectedRemoveFromPool: false,
			ExpectedMessage:        "",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			ic, pr := checkTestConfig()
			c.FuncPre(pr)
			status, removeFromMergePool, msg := checkConditionsFull(ic.Spec.MergeConfig.Query, pr)
			assert.Equal(t, c.ExpectedResult, status)
			assert.Equal(t, c.ExpectedRemoveFromPool, removeFromMergePool)
			assert.Equal(t, c.ExpectedMessage, msg)
		})
	}
}

type checkBranchAuthorTestCase struct {
	Value string
	Query cicdv1.MergeQuery

	ExpectedResult  bool
	ExpectedMessage string
}

func TestCheckBranch(t *testing.T) {
	tc := map[string]checkBranchAuthorTestCase{
		"success": {
			Value: "master",
			Query: cicdv1.MergeQuery{
				Branches: []string{"master", "master2"},
			},
			ExpectedResult:  true,
			ExpectedMessage: "",
		},
		"fail": {
			Value: "masters",
			Query: cicdv1.MergeQuery{
				Branches: []string{"master", "master2"},
			},
			ExpectedResult:  false,
			ExpectedMessage: "Branch [masters] is not in branches query.",
		},
		"failSkipBranch": {
			Value: "master",
			Query: cicdv1.MergeQuery{
				SkipBranches: []string{"master", "master2"},
			},
			ExpectedResult:  false,
			ExpectedMessage: "Branch [master] is in skipBranches query.",
		},
		"successSkip": {
			Value: "masters",
			Query: cicdv1.MergeQuery{
				SkipBranches: []string{"master", "master2"},
			},
			ExpectedResult:  true,
			ExpectedMessage: "",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			result, msg := checkBranch(c.Value, c.Query)
			assert.Equal(t, c.ExpectedResult, result)
			assert.Equal(t, c.ExpectedMessage, msg)
		})
	}
}

func TestCheckAuthor(t *testing.T) {
	tc := map[string]checkBranchAuthorTestCase{
		"success": {
			Value: "cqbqdd11519",
			Query: cicdv1.MergeQuery{
				Authors: []string{"cqbqdd11519"},
			},
			ExpectedResult:  true,
			ExpectedMessage: "",
		},
		"fail": {
			Value: "sunghyunkim3",
			Query: cicdv1.MergeQuery{
				Authors: []string{"cqbqdd11519"},
			},
			ExpectedResult:  false,
			ExpectedMessage: "Author [sunghyunkim3] is not in authors query.",
		},
		"failSkip": {
			Value: "cqbqdd11519",
			Query: cicdv1.MergeQuery{
				SkipAuthors: []string{"cqbqdd11519"},
			},
			ExpectedResult:  false,
			ExpectedMessage: "Author [cqbqdd11519] is in skipAuthors query.",
		},
		"successSkip": {
			Value: "sunghyunkim3",
			Query: cicdv1.MergeQuery{
				SkipAuthors: []string{"cqbqdd11519"},
			},
			ExpectedResult:  true,
			ExpectedMessage: "",
		},
		"successNone": {
			Value:           "sunghyunkim3",
			Query:           cicdv1.MergeQuery{},
			ExpectedResult:  true,
			ExpectedMessage: "",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			result, msg := checkAuthor(c.Value, c.Query)
			assert.Equal(t, c.ExpectedResult, result)
			assert.Equal(t, c.ExpectedMessage, msg)
		})
	}
}

type checkLabelsTestCase struct {
	Labels          map[string]struct{}
	Query           cicdv1.MergeQuery
	ExpectedResult  bool
	ExpectedMessage string
}

func TestCheckLabels(t *testing.T) {
	tc := map[string]checkLabelsTestCase{
		"success": {
			Labels: map[string]struct{}{
				"lgtm": {},
				"hold": {},
			},
			Query: cicdv1.MergeQuery{
				Labels:      []string{},
				BlockLabels: []string{},
			},
			ExpectedResult:  true,
			ExpectedMessage: "",
		},
		"failLabel": {
			Labels: nil,
			Query: cicdv1.MergeQuery{
				Labels: []string{
					"lgtm",
				},
				BlockLabels: []string{
					"hold",
				},
			},
			ExpectedResult:  false,
			ExpectedMessage: "Label [lgtm] is required.",
		},
		"successLgtm": {
			Labels: map[string]struct{}{
				"lgtm": {},
			},
			Query: cicdv1.MergeQuery{
				Labels: []string{
					"lgtm",
				},
				BlockLabels: []string{
					"hold",
				},
			},
			ExpectedResult:  true,
			ExpectedMessage: "",
		},
		"failHold": {
			Labels: map[string]struct{}{
				"lgtm": {},
				"hold": {},
			},
			Query: cicdv1.MergeQuery{
				Labels: []string{
					"lgtm",
				},
				BlockLabels: []string{
					"hold",
				},
			},
			ExpectedResult:  false,
			ExpectedMessage: "Label [hold] is blocking the merge.",
		},
		"failApproved": {
			Labels: map[string]struct{}{
				"lgtm":     {},
				"kind/bug": {},
			},
			Query: cicdv1.MergeQuery{
				Labels: []string{
					"lgtm",
					"approved",
				},
				BlockLabels: []string{
					"hold",
				},
			},
			ExpectedResult:  false,
			ExpectedMessage: "Label [approved] is required.",
		},
		"failBoth": {
			Labels: map[string]struct{}{
				"hold": {},
			},
			Query: cicdv1.MergeQuery{
				Labels: []string{
					"lgtm",
					"approved",
				},
				BlockLabels: []string{
					"hold",
				},
			},
			ExpectedResult:  false,
			ExpectedMessage: "Label [approved,lgtm] is required. Label [hold] is blocking the merge.",
		},
		"failHold2": {
			Labels: map[string]struct{}{
				"lgtm":     {},
				"approved": {},
				"hold":     {},
			},
			Query: cicdv1.MergeQuery{
				Labels: []string{
					"lgtm",
					"approved",
				},
				BlockLabels: []string{
					"hold",
				},
			},
			ExpectedResult:  false,
			ExpectedMessage: "Label [hold] is blocking the merge.",
		},
		"successBoth": {
			Labels: map[string]struct{}{
				"lgtm":     {},
				"approved": {},
			},
			Query: cicdv1.MergeQuery{
				Labels: []string{
					"lgtm",
					"approved",
				},
				BlockLabels: []string{
					"hold",
				},
			},
			ExpectedResult:  true,
			ExpectedMessage: "",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			result, msg := checkLabels(c.Labels, c.Query)

			assert.Equal(t, c.ExpectedResult, result, "Result")
			assert.Equal(t, c.ExpectedMessage, msg, "Message")
		})
	}
}

type checkChecksTestCase struct {
	Statuses        map[string]git.CommitStatus
	Query           cicdv1.MergeQuery
	ExpectedResult  bool
	ExpectedMessage string
}

func TestCheckChecks(t *testing.T) {
	tc := map[string]checkChecksTestCase{
		"failLint": {
			Statuses: map[string]git.CommitStatus{
				"test-lint": {State: "pending"},
				"test-unit": {State: "success"},
			},
			Query: cicdv1.MergeQuery{
				Checks: []string{},
			},
			ExpectedResult:  false,
			ExpectedMessage: "Checks [test-lint] are not successful.",
		},
		"successOptional": {
			Statuses: map[string]git.CommitStatus{
				"test-lint": {State: "pending"},
				"test-unit": {State: "success"},
			},
			Query: cicdv1.MergeQuery{
				OptionalChecks: []string{
					"test-lint",
				},
			},
			ExpectedResult:  true,
			ExpectedMessage: "",
		},
		"successRequired": {
			Statuses: map[string]git.CommitStatus{
				"test-lint": {State: "pending"},
				"test-unit": {State: "success"},
			},
			Query: cicdv1.MergeQuery{
				Checks: []string{
					"test-unit",
				},
			},
			ExpectedResult:  true,
			ExpectedMessage: "",
		},
		"failRequired": {
			Statuses: map[string]git.CommitStatus{
				"test-lint": {State: "pending"},
				"test-unit": {State: "pending"},
			},
			Query: cicdv1.MergeQuery{
				Checks: []string{
					"test-unit",
				},
			},
			ExpectedResult:  false,
			ExpectedMessage: "Checks [test-unit] are not successful.",
		},
		"failDefault": {
			Statuses: map[string]git.CommitStatus{
				"test-lint": {State: "pending"},
				"test-unit": {State: "pending"},
			},
			Query:           cicdv1.MergeQuery{},
			ExpectedResult:  false,
			ExpectedMessage: "Checks [test-lint,test-unit] are not successful.",
		},
		"failRequired2": {
			Statuses: map[string]git.CommitStatus{
				"test-lint": {State: "pending"},
				"test-unit": {State: "pending"},
			},
			Query: cicdv1.MergeQuery{
				Checks: []string{
					"test-lint",
					"test-unit",
				},
			},
			ExpectedResult:  false,
			ExpectedMessage: "Checks [test-lint,test-unit] are not successful.",
		},
		"success": {
			Statuses: map[string]git.CommitStatus{
				blockerContext: {State: "pending"},
				"test-lint":    {State: "pending"},
				"test-unit":    {State: "success"},
			},
			Query: cicdv1.MergeQuery{
				OptionalChecks: []string{
					"test-lint",
				},
			},
			ExpectedResult:  true,
			ExpectedMessage: "",
		},
		"failPassBlocker": {
			Statuses: map[string]git.CommitStatus{
				blockerContext: {State: "pending"},
				"test-lint":    {State: "pending"},
			},
			Query: cicdv1.MergeQuery{
				Checks: []string{
					"test-unit",
				},
			},
			ExpectedResult:  false,
			ExpectedMessage: "Checks [test-unit] are not successful.",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			result, msg := checkChecks(c.Statuses, c.Query)

			assert.Equal(t, c.ExpectedResult, result, "Result")
			assert.Equal(t, c.ExpectedMessage, msg, "Message")
		})
	}
}

func checkTestConfig() (*cicdv1.IntegrationConfig, *PullRequest) {
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

	pr := &PullRequest{
		PullRequest: git.PullRequest{
			ID:     testPRID,
			Head:   git.Head{Sha: testSHA},
			Labels: []git.IssueLabel{{Name: "kind/bug"}},
		},
		Statuses: map[string]git.CommitStatus{},
	}

	return ic, pr
}
