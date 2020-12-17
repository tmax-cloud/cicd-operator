package pipelinemanager

import (
	"context"
	"fmt"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateApprovalRunTask(job *cicdv1.IntegrationJob, j cicdv1.Job, ns string, c client.Client, task *tektonv1beta1.PipelineTask) error {
	task.TaskRef = generateCustomTaskRef(cicdv1.CustomTaskKindApproval)

	// Get approvers
	approvers := j.Approval.Approvers
	// from configMap
	if j.Approval.ApproversConfigMap != nil {
		cm := &corev1.ConfigMap{}
		if err := c.Get(context.Background(), types.NamespacedName{Name: j.Approval.ApproversConfigMap.Name, Namespace: ns}, cm); err != nil {
			return err
		}
		cmListStr, exist := cm.Data[cicdv1.CustomTaskApprovalApproversConfigMapKey]
		if !exist {
			return fmt.Errorf("there is no key '%s' in configmap %s/%s", cicdv1.CustomTaskApprovalApproversConfigMapKey, cm.Namespace, cm.Name)
		}
		cmList, err := utils.ParseApproversList(cmListStr)
		if err != nil {
			return err
		}
		approvers = append(approvers, cmList...)
	}

	// Get message
	msg := j.Approval.RequestMessage

	var sender, link string
	if job.Spec.Refs.Pull != nil {
		sender = job.Spec.Refs.Pull.Author.Name
		link = job.Spec.Refs.Pull.Link
	} else {
		// TODO: Sender for Push event
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

	return nil
}
