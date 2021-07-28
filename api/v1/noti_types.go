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

// Notification specifies notification
type Notification struct {
	// OnSuccess notifies when the job is succeeded
	OnSuccess *NotificationMethods `json:"onSuccess,omitempty"`

	// OnFailure notifies when the job is failed
	OnFailure *NotificationMethods `json:"onFailure,omitempty"`
}

// NotificationMethods specifies notification methods
type NotificationMethods struct {
	// Email sends email
	Email *NotiEmail `json:"email,omitempty"`

	// Slack sends slack
	Slack *NotiSlack `json:"slack,omitempty"`
}

// NotiEmail sends email to receivers
type NotiEmail struct {
	// Receivers is a list of email receivers
	Receivers []string `json:"receivers,omitempty"`

	// Title of the email
	Title string `json:"title"`

	// Content of the email
	Content string `json:"content"`

	// IsHTML describes if it's html content. Default is false
	IsHTML bool `json:"isHtml,omitempty"`
}

// NotiSlack sends slack to the webhook
type NotiSlack struct {
	// URL is a webhook url of a slack app. Refer to https://api.slack.com/messaging/webhooks
	URL string `json:"url"`

	// Message is a message sent to the webhook. It should be a Markdown format.
	// You can use $INTEGRATION_JOB_NAME and $JOB_NAME variable for IntegrationJob's name and the job's name respectively.
	Message string `json:"message"`
}
