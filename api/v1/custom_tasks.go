package v1

import "fmt"

const (
	CustomTaskGroup   = "cicd.tmax.io"
	CustomTaskVersion = "v1"
)

// Approval
const (
	CustomTaskKindApproval = "Approval"

	CustomTaskApprovalParamKeyApprovers         = "approvers"
	CustomTaskApprovalParamKeyMessage           = "message"
	CustomTaskApprovalParamKeyIntegrationJob    = "integration-job"
	CustomTaskApprovalParamKeyIntegrationJobJob = "integration-job-job"
	CustomTaskApprovalParamKeySender            = "sender"
	CustomTaskApprovalParamKeyLink              = "link"

	CustomTaskApprovalApproversConfigMapKey = "approvers"
)

var CustomTaskApiVersion = fmt.Sprintf("%s/%s", CustomTaskGroup, CustomTaskVersion)
