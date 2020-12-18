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
	CustomTaskApprovalParamKeySender            = "sender"
	CustomTaskApprovalParamKeyLink              = "link"

	CustomTaskApprovalApproversConfigMapKey = "approvers"
)

var CustomTaskApiVersion = fmt.Sprintf("%s/%s", CustomTaskGroup, CustomTaskVersion)
