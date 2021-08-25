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

package github

// UserInfo is a body of user get API
type UserInfo struct {
	ID       int    `json:"id"`
	UserName string `json:"login"`
	Email    string `json:"email"`
}

// UserPermission is a user's permission on a repository
type UserPermission struct {
	Permission string `json:"permission"`
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
	Context     string `json:"context"`
	State       string `json:"state"`
	Description string `json:"description"`
	TargetURL   string `json:"target_url"`
}

// CommentBody is a body structure for creating new comment
type CommentBody struct {
	Body string `json:"body"`
}

// LabelBody is a body structure for setting a label to issues/prs
type LabelBody struct {
	Name string `json:"name"`
}

// BranchResponse is a respond struct for branch request
type BranchResponse struct {
	Name   string `json:"name"`
	Commit struct {
		Sha string `json:"sha"`
	} `json:"commit"`
}

// MergeRequest is a request struct to merge a pull request
type MergeRequest struct {
	CommitTitle   string `json:"commit_title,omitempty"`
	CommitMessage string `json:"commit_message,omitempty"`
	MergeMethod   string `json:"merge_method"`
	Sha           string `json:"sha"`
}

// DiffFiles is a list of DiffFile
type DiffFiles []DiffFile

// DiffFile is a
type DiffFile struct {
	Filename     string `json:"filename"`
	PrevFilename string `json:"previous_filename"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	Changes      int    `json:"changes"`
}

// CommitResponse is a commits list response
type CommitResponse struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
		Committer struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"committer"`
	} `json:"commit"`
}
