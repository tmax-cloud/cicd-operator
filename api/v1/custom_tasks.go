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

import "fmt"

// API group and version for the custom tasks
const (
	CustomTaskGroup   = "cicd.tmax.io"
	CustomTaskVersion = "v1"
)

// Approval custom tasks
const (
	CustomTaskKindApproval = "ApprovalTask"

	CustomTaskApprovalParamKeyApprovers         = "approvers"
	CustomTaskApprovalParamKeyApproversCM       = "approvers-config-map"
	CustomTaskApprovalParamKeyMessage           = "message"
	CustomTaskApprovalParamKeyIntegrationJob    = "integration-job"
	CustomTaskApprovalParamKeyIntegrationJobJob = "integration-job-job"
	CustomTaskApprovalParamKeySenderName        = "senderName"
	CustomTaskApprovalParamKeySenderEmail       = "senderEmail"
	CustomTaskApprovalParamKeyLink              = "link"

	CustomTaskApprovalApproversConfigMapKey = "approvers"
)

// Email custom tasks
const (
	CustomTaskKindEmail = "EmailTask"

	CustomTaskEmailParamKeyReceivers         = "receivers"
	CustomTaskEmailParamKeyTitle             = "title"
	CustomTaskEmailParamKeyContent           = "content"
	CustomTaskEmailParamKeyIsHTML            = "isHtml"
	CustomTaskEmailParamKeyIntegrationJobJob = CustomTaskApprovalParamKeyIntegrationJobJob
	CustomTaskEmailParamKeyIntegrationJob    = CustomTaskApprovalParamKeyIntegrationJob
)

// Slack custom tasks
const (
	CustomTaskKindSlack = "SlackTask"

	CustomTaskSlackParamKeyWebhook           = "webhook-url"
	CustomTaskSlackParamKeyMessage           = CustomTaskApprovalParamKeyMessage
	CustomTaskSlackParamKeyIntegrationJob    = CustomTaskApprovalParamKeyIntegrationJob
	CustomTaskSlackParamKeyIntegrationJobJob = CustomTaskApprovalParamKeyIntegrationJobJob
)

// CustomTaskAPIVersion is
var CustomTaskAPIVersion = fmt.Sprintf("%s/%s", CustomTaskGroup, CustomTaskVersion)
