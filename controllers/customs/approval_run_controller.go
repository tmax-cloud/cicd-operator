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

package customs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ApprovalRunHandler handles approval custom task
type ApprovalRunHandler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Handle generates Approval
func (a *ApprovalRunHandler) Handle(run *tektonv1alpha1.Run) (ctrl.Result, error) {
	ctx := context.Background()
	log := a.Log.WithValues("ApprovalRun", run.Namespace)

	original := run.DeepCopy()

	// New Condition default
	cond := run.Status.GetCondition(apis.ConditionSucceeded)
	if cond == nil {
		cond = &apis.Condition{
			Type:    apis.ConditionSucceeded,
			Status:  corev1.ConditionUnknown,
			Reason:  "Awaiting",
			Message: "waiting for approval",
		}
	}

	defer func() {
		run.Status.SetCondition(cond)
		p := client.MergeFrom(original)
		if err := a.Client.Status().Patch(ctx, run, p); err != nil {
			log.Error(err, "")
		}
	}()

	oldCond := run.Status.GetCondition(apis.ConditionSucceeded)
	// Ignore if it's completed (has completionTime)
	if run.Status.CompletionTime != nil || (oldCond != nil && oldCond.Status != corev1.ConditionUnknown) {
		return ctrl.Result{}, nil
	}

	// Set startTime
	if run.Status.StartTime == nil {
		run.Status.StartTime = &metav1.Time{Time: time.Now()}
	}

	// Check approval
	approval, err := a.newApproval(run)
	if err != nil {
		log.Error(err, "")
		a.setApprovalRunStatus(cond, corev1.ConditionFalse, "CannotCreateApproval", err.Error())
		return ctrl.Result{}, nil
	}

	if err := a.Client.Get(context.Background(), types.NamespacedName{Name: approval.Name, Namespace: approval.Namespace}, approval); err != nil {
		if errors.IsNotFound(err) {
			// Create approval if not exist
			if err := controllerutil.SetControllerReference(run, approval, a.Scheme); err != nil {
				log.Error(err, "")
				a.setApprovalRunStatus(cond, corev1.ConditionFalse, "CannotCreateApproval", err.Error())
				return ctrl.Result{}, nil
			}
			if err := a.Client.Create(ctx, approval); err != nil {
				log.Error(err, "")
				a.setApprovalRunStatus(cond, corev1.ConditionFalse, "CannotCreateApproval", err.Error())
				return ctrl.Result{}, nil
			}

			return ctrl.Result{}, nil
		}
		a.setApprovalRunStatus(cond, corev1.ConditionFalse, "ErrorGettingApproval", err.Error())
		return ctrl.Result{}, nil
	}

	// Reflect approval status to Run
	if approval.Status.DecisionTime != nil {
		var reason = string(approval.Status.Result)
		if approval.Status.Result == cicdv1.ApprovalResultApproved {
			cond.Status = corev1.ConditionTrue
		} else {
			cond.Status = corev1.ConditionFalse
		}
		cond.Reason = reason
		cond.Message = fmt.Sprintf("%s %s this approval, reason: %s, decisionTime: %s", approval.Status.Approver, strings.ToLower(reason), approval.Status.Reason, approval.Status.DecisionTime)
		run.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	} else {
		cond.Status = corev1.ConditionUnknown
		a.reflectApprovalEmailStatus(approval, cond)
	}

	return ctrl.Result{}, nil
}

func (a *ApprovalRunHandler) reflectApprovalEmailStatus(approval *cicdv1.Approval, cond *apis.Condition) {
	// Reflect approval email status to Run
	// Result of sending email is not critical to the Approval itself, so it's only stated in the message
	var mailStatuses []string

	sentReqMail := approval.Status.Conditions.GetCondition(cicdv1.ApprovalConditionSentRequestMail)
	if sentReqMail != nil && sentReqMail.IsFalse() {
		mailStatuses = append(mailStatuses, fmt.Sprintf("RequestMail : %s-%s", sentReqMail.Reason, sentReqMail.Message))
	}

	sentResMail := approval.Status.Conditions.GetCondition(cicdv1.ApprovalConditionSentResultMail)
	if sentResMail != nil && sentResMail.IsFalse() {
		mailStatuses = append(mailStatuses, fmt.Sprintf("ResultMail : %s-%s", sentResMail.Reason, sentResMail.Message))
	}

	cond.Message = strings.Join(mailStatuses, ", ")
}

func (a *ApprovalRunHandler) setApprovalRunStatus(cond *apis.Condition, status corev1.ConditionStatus, reason, message string) {
	if cond == nil {
		return
	}

	cond.Status = status
	cond.Reason = reason
	cond.Message = message
}

func (a *ApprovalRunHandler) newApproval(run *tektonv1alpha1.Run) (*cicdv1.Approval, error) {
	prName := "none"

	// Get PipelineRun (owner)
	for _, owner := range run.OwnerReferences {
		if strings.HasPrefix(owner.APIVersion, "tekton.dev/v1") && owner.Kind == "PipelineRun" {
			prName = owner.Name
			break
		}
	}

	// Get parameters
	_, approvers, err := searchParam(run.Spec.Params, cicdv1.CustomTaskApprovalParamKeyApprovers, tektonv1beta1.ParamTypeArray)
	if err != nil {
		return nil, err
	}

	approverCm, _, err := searchParam(run.Spec.Params, cicdv1.CustomTaskApprovalParamKeyApproversCM, tektonv1beta1.ParamTypeString)
	if err != nil {
		return nil, err
	}
	if approverCm != "" {
		cm := &corev1.ConfigMap{}
		_ = a.Client.Get(context.Background(), types.NamespacedName{Name: approverCm, Namespace: run.Namespace}, cm)
		if cm.Data != nil {
			cmListStr, exist := cm.Data[cicdv1.CustomTaskApprovalApproversConfigMapKey]
			if exist {
				cmList, _ := utils.ParseApproversList(cmListStr)
				if cmList != nil {
					approvers = append(approvers, cmList...)
				}
			}
		}
	}

	msg, _, err := searchParam(run.Spec.Params, cicdv1.CustomTaskApprovalParamKeyMessage, tektonv1beta1.ParamTypeString)
	if err != nil {
		return nil, err
	}

	jobName, _, err := searchParam(run.Spec.Params, cicdv1.CustomTaskApprovalParamKeyIntegrationJob, tektonv1beta1.ParamTypeString)
	if err != nil {
		return nil, err
	}

	jobJobName, _, err := searchParam(run.Spec.Params, cicdv1.CustomTaskApprovalParamKeyIntegrationJobJob, tektonv1beta1.ParamTypeString)
	if err != nil {
		return nil, err
	}

	sender, _, err := searchParam(run.Spec.Params, cicdv1.CustomTaskApprovalParamKeySenderName, tektonv1beta1.ParamTypeString)
	if err != nil {
		return nil, err
	}

	// Sender email can be empty
	senderEmail, _, _ := searchParam(run.Spec.Params, cicdv1.CustomTaskApprovalParamKeySenderEmail, tektonv1beta1.ParamTypeString)

	link, _, err := searchParam(run.Spec.Params, cicdv1.CustomTaskApprovalParamKeyLink, tektonv1beta1.ParamTypeString)
	if err != nil {
		return nil, err
	}

	return &cicdv1.Approval{
		ObjectMeta: metav1.ObjectMeta{
			Name:      run.Name,
			Namespace: run.Namespace,
		},
		Spec: cicdv1.ApprovalSpec{
			SkipSendMail:   !configs.EnableMail,
			PipelineRun:    prName,
			Users:          approvers,
			IntegrationJob: jobName,
			JobName:        jobJobName,
			Sender: &cicdv1.ApprovalSender{
				Name:  sender,
				Email: senderEmail,
			},
			Message: msg,
			Link:    link,
		},
	}, nil
}

func searchParam(params []tektonv1beta1.Param, key string, expectedKind tektonv1beta1.ParamType) (string, []string, error) {
	for _, p := range params {
		if p.Name == key {
			if p.Value.Type == expectedKind {
				switch p.Value.Type {
				case tektonv1beta1.ParamTypeString:
					return p.Value.StringVal, nil, nil
				case tektonv1beta1.ParamTypeArray:
					return "", p.Value.ArrayVal, nil
				}
			} else {
				return "", nil, fmt.Errorf("parameter %s is expected to be %s ,got %s", key, expectedKind, p.Value.Type)
			}
		}
	}

	return "", nil, fmt.Errorf("there is no param %s for Run", key)
}
