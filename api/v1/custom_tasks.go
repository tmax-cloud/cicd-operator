package v1

import "fmt"

const (
	CustomTaskGroup   = "cicd.tmax.io"
	CustomTaskVersion = "v1"
)

// Approval
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

// Email
const (
	CustomTaskKindEmail = "EmailTask"

	CustomTaskEmailParamKeyReceivers = "receivers"
	CustomTaskEmailParamKeyTitle     = "title"
	CustomTaskEmailParamKeyContent   = "content"
	CustomTaskEmailParamKeyIsHtml    = "isHtml"
)

var CustomTaskApiVersion = fmt.Sprintf("%s/%s", CustomTaskGroup, CustomTaskVersion)
