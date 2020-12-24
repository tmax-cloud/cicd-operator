package pipelinemanager

import (
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
)

func generateApprovalRunTask(job *cicdv1.IntegrationJob, j cicdv1.Job, task *tektonv1beta1.PipelineTask) error {
	task.TaskRef = generateCustomTaskRef(cicdv1.CustomTaskKindApproval)

	// Get approvers
	approvers := j.Approval.Approvers

	// Get message
	msg := j.Approval.RequestMessage

	// Get sender
	sender := job.Spec.Refs.Sender

	var link string
	if job.Spec.Refs.Pull != nil {
		link = job.Spec.Refs.Pull.Link
	} else {
		link = job.Spec.Refs.Base.Link
	}

	task.Params = append(task.Params, []tektonv1beta1.Param{
		{Name: cicdv1.CustomTaskApprovalParamKeyApprovers, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: approvers}},
		{Name: cicdv1.CustomTaskApprovalParamKeyMessage, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: msg}},
		{Name: cicdv1.CustomTaskApprovalParamKeyIntegrationJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: job.Name}},
		{Name: cicdv1.CustomTaskApprovalParamKeyIntegrationJobJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: j.Name}},
		{Name: cicdv1.CustomTaskApprovalParamKeySender, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: sender}},
		{Name: cicdv1.CustomTaskApprovalParamKeyLink, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: link}},
	}...)

	if j.Approval.ApproversConfigMap != nil && j.Approval.ApproversConfigMap.Name != "" {
		task.Params = append(task.Params, tektonv1beta1.Param{Name: cicdv1.CustomTaskApprovalParamKeyApproversCM, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: j.Approval.ApproversConfigMap.Name}})
	}

	return nil
}
