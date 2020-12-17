/*


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

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-lib/status"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/mail"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
)

// ApprovalReconciler reconciles a Approval object
type ApprovalReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cicd.tmax.io,resources=approvals,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cicd.tmax.io,resources=approvals/status,verbs=get;update;patch

func (r *ApprovalReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("approval", req.NamespacedName)

	// Get Approval
	instance := &cicdv1.Approval{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		log.Error(err, "")
		return reconcile.Result{}, err
	}

	sentReqMailCond := instance.Status.Conditions.GetCondition(cicdv1.ApprovalConditionSentRequestMail)
	if sentReqMailCond == nil {
		sentReqMailCond = &status.Condition{
			Type:   cicdv1.ApprovalConditionSentRequestMail,
			Status: corev1.ConditionUnknown,
		}
	}

	sentResMailCond := instance.Status.Conditions.GetCondition(cicdv1.ApprovalConditionSentResultMail)
	if sentResMailCond == nil {
		sentResMailCond = &status.Condition{
			Type:   cicdv1.ApprovalConditionSentResultMail,
			Status: corev1.ConditionUnknown,
		}
	}

	defer func() {
		instance.Status.Conditions.SetCondition(*sentReqMailCond)
		instance.Status.Conditions.SetCondition(*sentResMailCond)
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			log.Error(err, "")
		}
	}()

	// Default conditions
	if instance.Status.Result == "" {
		instance.Status.Result = cicdv1.ApprovalResultWaiting
	}

	// Set SentRequestMail
	if sentReqMailCond == nil || sentReqMailCond.Status == corev1.ConditionUnknown {
		title := "[CI/CD] Approval is requested to you"
		content := fmt.Sprintf("IntegrationJob: %s\nJobName: %s\nLink: %s\nMessage: %s", instance.Spec.IntegrationJob, instance.Spec.JobName, instance.Spec.Link, instance.Spec.Message)
		r.sendMail(instance.Spec.Users, title, content, sentReqMailCond)
	}

	// Set SentResultMail - only if is decided
	if instance.Status.Result == cicdv1.ApprovalResultApproved || instance.Status.Result == cicdv1.ApprovalResultRejected {
		title := fmt.Sprintf("[CI/CD] Approval is %s", strings.ToLower(string(instance.Status.Result)))
		content := fmt.Sprintf("IntegrationJob: %s\nJobName: %s\nLink: %s\nMessage: %s\n\nResult\nResult: %s\nReason: %s\nDecisionTime: %s", instance.Spec.IntegrationJob, instance.Spec.JobName, instance.Spec.Link, instance.Spec.Message, instance.Status.Result, instance.Status.Reason, instance.Status.DecisionTime)
		if sentResMailCond == nil || sentResMailCond.Status == corev1.ConditionUnknown {
			r.sendMail(instance.Spec.Users, title, content, sentResMailCond)
		}
	}

	return ctrl.Result{}, nil
}

func (r *ApprovalReconciler) sendMail(users []string, title, content string, cond *status.Condition) {
	sentMail := false
	cond.Reason = "EmailDisabled"
	if configs.EnableMail {
		tos := utils.ParseEmailFromUsers(users)
		if err := mail.Send(tos, title, content, false, r.Client); err != nil {
			r.Log.Error(err, "")
			cond.Reason = "ErrorSendingMail"
			cond.Message = err.Error()
		} else {
			sentMail = true
		}
	}

	if sentMail {
		cond.Status = corev1.ConditionTrue
		cond.Reason = ""
		cond.Message = ""
	} else {
		cond.Status = corev1.ConditionFalse
	}
}

func (r *ApprovalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cicdv1.Approval{}).
		Complete(r)
}
