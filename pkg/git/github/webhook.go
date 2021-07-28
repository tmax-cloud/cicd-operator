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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// PullRequestWebhook is a github-specific pull-request event webhook body
type PullRequestWebhook struct {
	Action string `json:"action"`
	Number int    `json:"number"`
	Sender User   `json:"sender"`

	PullRequest PullRequest `json:"pull_request"`

	Repo Repo `json:"repository"`

	// Changed label
	Label LabelBody `json:"label"`
}

// PushWebhook is a github-specific push event webhook body
type PushWebhook struct {
	Ref    string `json:"ref"`
	Repo   Repo   `json:"repository"`
	Sender User   `json:"sender"`
	Sha    string `json:"after"`
}

// IssueCommentWebhook is a github-specific issue_comment webhook body
type IssueCommentWebhook struct {
	Action  string  `json:"action"`
	Comment Comment `json:"comment"`
	Issue   struct {
		PullRequest struct {
			URL string `json:"url"`
		} `json:"pull_request"`
	} `json:"issue"`
	Repo   Repo `json:"repository"`
	Sender User `json:"sender"`
}

// PullRequestReviewWebhook is a github-specific pull_request_review webhook body
type PullRequestReviewWebhook struct {
	Action string `json:"action"`
	Review struct {
		Body        string       `json:"body"`
		SubmittedAt *metav1.Time `json:"submitted_at"`
		State       string       `json:"state"`
		User        User         `json:"user"`
	} `json:"review"`
	PullRequest PullRequest `json:"pull_request"`
	Repo        Repo        `json:"repository"`
	Sender      User        `json:"sender"`
}

// PullRequestReviewCommentWebhook is a github-specific pull_request_review_comment webhook body
type PullRequestReviewCommentWebhook struct {
	Action      string      `json:"action"`
	Comment     Comment     `json:"comment"`
	PullRequest PullRequest `json:"pull_request"`
	Repo        Repo        `json:"repository"`
	Sender      User        `json:"sender"`
}

// Repo structure for webhook event
type Repo struct {
	Name  string `json:"full_name"`
	URL   string `json:"html_url"`
	Owner struct {
		ID string `json:"login"`
	} `json:"owner"`
	Private bool `json:"private"`
}

// PullRequest is a pull request info
type PullRequest struct {
	Title     string `json:"title"`
	Number    int    `json:"number"`
	State     string `json:"state"`
	URL       string `json:"html_url"`
	Mergeable bool   `json:"mergeable"`
	User      User   `json:"user"`
	Draft     bool   `json:"draft"`
	Head      struct {
		Ref string `json:"ref"`
		Sha string `json:"sha"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
		Sha string `json:"sha"`
	} `json:"base"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

// User is a sender of the event
type User struct {
	Name string `json:"login"`
	ID   int    `json:"id"`
}

// Comment is a comment payload
type Comment struct {
	Body      string       `json:"body"`
	User      User         `json:"user"`
	CreatedAt *metav1.Time `json:"created_at"`
	UpdatedAt *metav1.Time `json:"updated_at"`
}

// RegistrationWebhookBody is a request body for registering webhook to remote git server
type RegistrationWebhookBody struct {
	Name   string                        `json:"name"`
	Active bool                          `json:"active"`
	Events []string                      `json:"events"`
	Config RegistrationWebhookBodyConfig `json:"config"`
}

// RegistrationWebhookBodyConfig is a config for the webhook
type RegistrationWebhookBodyConfig struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	InsecureSsl string `json:"insecure_ssl"`
	Secret      string `json:"secret"`
}

// WebhookEntry is a body of list of registered webhooks
type WebhookEntry struct {
	ID     int `json:"id"`
	Config struct {
		URL string `json:"url"`
	} `json:"config"`
}
