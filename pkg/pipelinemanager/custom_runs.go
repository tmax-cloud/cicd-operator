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

package pipelinemanager

import (
	"strconv"

	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
)

// Approval custom tasks
func generateApprovalRunTask(job *cicdv1.IntegrationJob, j *cicdv1.Job, task *tektonv1beta1.PipelineTask) {
	task.TaskRef = generateCustomTaskRef(cicdv1.CustomTaskKindApproval)

	// Get approvers
	var approvers []string
	for _, approver := range j.Approval.Approvers {
		param := approver.Name
		if approver.Email != "" {
			param += "=" + approver.Email
		}
		approvers = append(approvers, param)
	}

	// Get message
	msg := j.Approval.RequestMessage

	// Get sender
	sender := job.Spec.Refs.Sender

	var link string
	if job.Spec.Refs.Pulls != nil {
		link = job.Spec.Refs.Pulls[0].Link
	} else {
		link = job.Spec.Refs.Base.Link
	}

	task.Params = append(task.Params, []tektonv1beta1.Param{
		{Name: cicdv1.CustomTaskApprovalParamKeyApprovers, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: approvers}},
		{Name: cicdv1.CustomTaskApprovalParamKeyMessage, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: msg}},
		{Name: cicdv1.CustomTaskApprovalParamKeyIntegrationJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: job.Name}},
		{Name: cicdv1.CustomTaskApprovalParamKeyIntegrationJobJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: j.Name}},
		{Name: cicdv1.CustomTaskApprovalParamKeySenderName, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: sender.Name}},
		{Name: cicdv1.CustomTaskApprovalParamKeySenderEmail, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: sender.Email}},
		{Name: cicdv1.CustomTaskApprovalParamKeyLink, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: link}},
	}...)

	approverCm := ""
	if j.Approval.ApproversConfigMap != nil && j.Approval.ApproversConfigMap.Name != "" {
		approverCm = j.Approval.ApproversConfigMap.Name
	}
	task.Params = append(task.Params, tektonv1beta1.Param{Name: cicdv1.CustomTaskApprovalParamKeyApproversCM, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: approverCm}})
}

// Email custom tasks
func generateEmailRunTask(job *cicdv1.IntegrationJob, j *cicdv1.Job, task *tektonv1beta1.PipelineTask) {
	task.TaskRef = generateCustomTaskRef(cicdv1.CustomTaskKindEmail)
	task.Params = append(task.Params, generateEmailRunParams(job, j, j.Email)...)
}

func generateEmailRunParams(job *cicdv1.IntegrationJob, j *cicdv1.Job, email *cicdv1.NotiEmail) []tektonv1beta1.Param {
	// Get receivers
	receivers := email.Receivers

	// Get title
	title := email.Title

	// Get content
	content := email.Content

	// Get isHTML
	isHTML := email.IsHTML

	return []tektonv1beta1.Param{
		{Name: cicdv1.CustomTaskEmailParamKeyReceivers, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: receivers}},
		{Name: cicdv1.CustomTaskEmailParamKeyTitle, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: title}},
		{Name: cicdv1.CustomTaskEmailParamKeyContent, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: content}},
		{Name: cicdv1.CustomTaskEmailParamKeyIsHTML, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: strconv.FormatBool(isHTML)}},
		{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: job.Name}},
		{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJobJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: j.Name}},
	}
}

// Slack custom tasks
func generateSlackRunTask(job *cicdv1.IntegrationJob, j *cicdv1.Job, task *tektonv1beta1.PipelineTask) {
	task.TaskRef = generateCustomTaskRef(cicdv1.CustomTaskKindSlack)
	task.Params = append(task.Params, generateSlackRunParams(job, j, j.Slack)...)
}

func generateSlackRunParams(job *cicdv1.IntegrationJob, j *cicdv1.Job, slack *cicdv1.NotiSlack) []tektonv1beta1.Param {
	return []tektonv1beta1.Param{
		{Name: cicdv1.CustomTaskSlackParamKeyWebhook, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: slack.URL}},
		{Name: cicdv1.CustomTaskSlackParamKeyMessage, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: slack.Message}},
		{Name: cicdv1.CustomTaskSlackParamKeyIntegrationJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: job.Name}},
		{Name: cicdv1.CustomTaskSlackParamKeyIntegrationJobJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: j.Name}},
	}
}

func generateCustomTaskRef(kind string) *tektonv1beta1.TaskRef {
	return &tektonv1beta1.TaskRef{
		APIVersion: cicdv1.CustomTaskAPIVersion,
		Kind:       tektonv1beta1.TaskKind(kind),
	}
}
