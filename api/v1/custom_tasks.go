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

	CustomTaskEmailParamKeyReceivers = "receivers"
	CustomTaskEmailParamKeyTitle     = "title"
	CustomTaskEmailParamKeyContent   = "content"
	CustomTaskEmailParamKeyIsHTML    = "isHtml"
)

// CustomTaskAPIVersion is
var CustomTaskAPIVersion = fmt.Sprintf("%s/%s", CustomTaskGroup, CustomTaskVersion)
