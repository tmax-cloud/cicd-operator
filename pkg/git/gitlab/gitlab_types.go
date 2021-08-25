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

package gitlab

// UserInfo is a body of user get API
type UserInfo struct {
	ID          int    `json:"id"`
	UserName    string `json:"username"`
	PublicEmail string `json:"public_email"`
	Email       string `json:"email"`
}

// UserPermission is a user's permission on a repository
type UserPermission struct {
	AccessLevel int `json:"access_level"`
}

// CommitStatusRequest is an API body for setting commits' status
type CommitStatusRequest struct {
	State       string `json:"state"`
	TargetURL   string `json:"target_url"`
	Description string `json:"description"`
	Context     string `json:"context"`
}

// CommitStatusResponse is a response body of getting commit status
type CommitStatusResponse struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
	TargetURL   string `json:"target_url"`
}

// CommentBody is a body structure for creating new comment
type CommentBody struct {
	Body string `json:"body"`
}

// UpdateMergeRequest is a struct to update a merge request
type UpdateMergeRequest struct {
	AddLabels    string `json:"add_labels"`
	RemoveLabels string `json:"remove_labels"`
}

// MergeRequest is a body struct of a merge request
type MergeRequest struct {
	ID     int    `json:"iid"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Author struct {
		ID       int    `json:"id"`
		UserName string `json:"username"`
	} `json:"author"`
	WebURL       string   `json:"web_url"`
	TargetBranch string   `json:"target_branch"`
	SourceBranch string   `json:"source_branch"`
	SHA          string   `json:"sha"`
	Labels       []string `json:"labels"`
	HasConflicts bool     `json:"has_conflicts"`
}

// BranchResponse is a respond struct for branch request
type BranchResponse struct {
	Name   string `json:"name"`
	Commit struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
}

// MergeAcceptRequest is a request struct to merge a merge request
type MergeAcceptRequest struct {
	MergeCommitMessage  string `json:"merge_commit_message,omitempty"`
	SquashCommitMessage string `json:"squash_commit_message,omitempty"`
	Squash              bool   `json:"squash"`
	Sha                 string `json:"sha"`
	RemoveSourceBranch  bool   `json:"should_remove_source_branch"`
}

// MergeRequestChanges is a changed list of the merge request
type MergeRequestChanges struct {
	Changes []struct {
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
		Diff    string `json:"diff"`
	} `json:"changes"`
}

// CommitResponse is a commits list response
type CommitResponse struct {
	ID             string `json:"id"`
	Message        string `json:"message"`
	AuthorName     string `json:"author_name"`
	AuthorEmail    string `json:"author_email"`
	CommitterName  string `json:"committer_name"`
	CommitterEmail string `json:"committer_email"`
}
