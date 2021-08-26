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

package v1

import "github.com/tmax-cloud/cicd-operator/pkg/git"

// MergeConfig is a config struct of the merge automation feature
type MergeConfig struct {
	// Method is a merge method
	// +kubebuilder:validation:Enum=squash;merge
	Method git.MergeMethod `json:"method,omitempty"`

	// CommitTemplate is a message template for a merge commit.
	// The commit message is compiled as a go template using blocker.PullRequest object.
	CommitTemplate string `json:"commitTemplate,omitempty"`

	// Query is conditions for a open PR to be merged
	Query MergeQuery `json:"query"`
}

// MergeQuery defines conditions for a open PR to be merged
type MergeQuery struct {
	// Labels specify the required labels of PR to be merged
	Labels []string `json:"labels,omitempty"`

	// BlockLabels specify the required labels of PR to be blocked for merge
	BlockLabels []string `json:"blockLabels,omitempty"`

	// Authors specify the required authors of PR to be merged
	// Authors and SkipAuthors are mutually exclusive
	Authors []string `json:"authors,omitempty"`

	// SkipAuthors specify the required authors of PR to be blocked for merge
	// Authors and SkipAuthors are mutually exclusive
	SkipAuthors []string `json:"skipAuthors,omitempty"`

	// Branches specify the required base branches of PR to be merged
	// Branches and SkipBranches are mutually exclusive
	Branches []string `json:"branches,omitempty"`

	// SkipBranches specify the required base branches of PR to be blocked for merge
	// Branches and SkipBranches are mutually exclusive
	SkipBranches []string `json:"skipBranches,omitempty"`

	// Checks are checks needed to be passed for the PR to be merged.
	// Checks and OptionalChecks are mutually exclusive
	Checks []string `json:"checks,omitempty"`

	// OptionalChecks are checks that are not required.
	// Checks and OptionalChecks are mutually exclusive
	OptionalChecks []string `json:"optionalChecks,omitempty"`

	// ApproveRequired specifies whether to check github/gitlab's approval
	ApproveRequired bool `json:"approveRequired,omitempty"`
}
