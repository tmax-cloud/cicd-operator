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

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"

	"github.com/go-logr/logr"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/notification/mail"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ApprovalReconciler reconciles a Approval object
type ApprovalReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	MailSender mail.Sender
}

const (
	// APIGroup is a group of api
	APIGroup           = "cicdapi.tmax.io"
	approvalKind       = "approvals"
	approvalSubApprove = "approve"
	approvalSubReject  = "reject"

	serviceAccountPrefix = "system:serviceaccount:"

	configMapPathEnv           = "EMAIL_TEMPLATE_PATH"
	configMapPathDefault       = "/templates/email"
	configMapKeyRequestTitle   = "request-title"
	configMapKeyRequestContent = "request-content"
	configMapKeyResultTitle    = "result-title"
	configMapKeyResultContent  = "result-content"
)

// +kubebuilder:rbac:groups=cicd.tmax.io,resources=approvals,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cicd.tmax.io,resources=approvals/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cicdapi.tmax.io,resources=approvals/approve;approvals/reject,verbs=update
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile reconciles Approval object
func (r *ApprovalReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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
	original := instance.DeepCopy()

	// Default conditions
	if sentReqMailCond := meta.FindStatusCondition(instance.Status.Conditions, cicdv1.ApprovalConditionSentRequestMail); sentReqMailCond == nil {
		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
			Type:    cicdv1.ApprovalConditionSentRequestMail,
			Status:  metav1.ConditionUnknown,
			Reason:  "NotProcessed",
			Message: "Request email is not processed yet",
		})
	}

	if sentResMailCond := meta.FindStatusCondition(instance.Status.Conditions, cicdv1.ApprovalConditionSentResultMail); sentResMailCond == nil {
		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
			Type:    cicdv1.ApprovalConditionSentResultMail,
			Status:  metav1.ConditionUnknown,
			Reason:  "NotProcessed",
			Message: "Result email is not processed yet",
		})
	}

	r.bumpV050(instance)

	defer func() {
		p := client.MergeFrom(original)
		if err := r.Client.Status().Patch(ctx, instance, p); err != nil {
			log.Error(err, "")
		}
	}()

	// Default conditions
	if instance.Status.Result == "" {
		instance.Status.Result = cicdv1.ApprovalResultAwaiting
		return ctrl.Result{}, nil
	}

	// Create Role/RoleBinding if not exists
	if err := r.createOrUpdateRole(instance); err != nil {
		log.Error(err, "updating roles")
		r.setError(instance, err.Error())
		return ctrl.Result{}, nil
	}
	if err := r.createOrUpdateRoleBinding(instance); err != nil {
		log.Error(err, "updating rolebindings")
		r.setError(instance, err.Error())
		return ctrl.Result{}, nil
	}

	// Process mails
	r.processMail(instance)
	return ctrl.Result{}, nil
}

// Update to v0.5.0 - reason, message became required
func (r *ApprovalReconciler) bumpV050(instance *cicdv1.Approval) {
	// Bump req cond
	if cond := meta.FindStatusCondition(instance.Status.Conditions, cicdv1.ApprovalConditionSentRequestMail); cond != nil {
		upgradeV050Condition(cond, "Sent", "NotSent")
	}

	// Bump res cond
	if cond := meta.FindStatusCondition(instance.Status.Conditions, cicdv1.ApprovalConditionSentResultMail); cond != nil {
		upgradeV050Condition(cond, "Sent", "NotSent")
	}
}

func (r *ApprovalReconciler) processMail(instance *cicdv1.Approval) {
	if instance.Spec.SkipSendMail {
		return
	}

	// Set SentRequestMail
	if reqCond := meta.FindStatusCondition(instance.Status.Conditions, cicdv1.ApprovalConditionSentRequestMail); reqCond == nil || reqCond.Status == "" || reqCond.Status == metav1.ConditionUnknown {
		title, content, err := r.generateMail(instance, configMapKeyRequestTitle, configMapKeyRequestContent)
		if err != nil {
			reqCond.Status = metav1.ConditionFalse
			reqCond.Reason = "EmailGenerateError"
			reqCond.Message = err.Error()
		} else {
			r.sendMail(utils.ParseEmailFromUsers(instance.Spec.Users), title, content, reqCond)
		}
	}

	// Set SentResultMail - only if is decided
	if instance.Status.Result != cicdv1.ApprovalResultApproved && instance.Status.Result != cicdv1.ApprovalResultRejected {
		return
	}
	if instance.Spec.Sender == nil {
		return
	}
	if resCond := meta.FindStatusCondition(instance.Status.Conditions, cicdv1.ApprovalConditionSentResultMail); resCond == nil || resCond.Status == metav1.ConditionUnknown {
		title, content, err := r.generateMail(instance, configMapKeyResultTitle, configMapKeyResultContent)
		if err != nil {
			resCond.Status = metav1.ConditionFalse
			resCond.Reason = "EmailGenerateError"
			resCond.Message = err.Error()
		} else {
			r.sendMail([]string{instance.Spec.Sender.Email}, title, content, resCond)
		}
	}
}

func (r *ApprovalReconciler) roleAndBindingName(approvalName string) string {
	return "cicd-approval-" + approvalName
}

func (r *ApprovalReconciler) createOrUpdateRole(approval *cicdv1.Approval) error {
	role := &rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.roleAndBindingName(approval.Name),
			Namespace: approval.Namespace,
		},
	}
	if err := r.Client.Get(context.Background(), types.NamespacedName{Name: role.Name, Namespace: role.Namespace}, role); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}
	original := role.DeepCopy()

	// Set labels, annotations
	role.Labels = r.labelsForRoleAndBinding(approval)
	role.Annotations = r.annotationsForRoleAndBinding(approval)

	// Set rules
	role.Rules = []rbac.PolicyRule{{
		APIGroups:     []string{APIGroup},
		Resources:     []string{fmt.Sprintf("%s/%s", approvalKind, approvalSubApprove), fmt.Sprintf("%s/%s", approvalKind, approvalSubReject)},
		ResourceNames: []string{approval.Name},
		Verbs:         []string{"update"},
	}}

	return utils.CreateOrPatchObject(role, original, approval, r.Client, r.Scheme)
}

func (r *ApprovalReconciler) createOrUpdateRoleBinding(approval *cicdv1.Approval) error {
	binding := &rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.roleAndBindingName(approval.Name),
			Namespace: approval.Namespace,
		},
	}
	if err := r.Client.Get(context.Background(), types.NamespacedName{Name: binding.Name, Namespace: binding.Namespace}, binding); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}
	original := binding.DeepCopy()

	// Set labels, annotations
	binding.Labels = r.labelsForRoleAndBinding(approval)
	binding.Annotations = r.annotationsForRoleAndBinding(approval)

	// Set bindings
	binding.RoleRef = rbac.RoleRef{
		APIGroup: rbac.GroupName,
		Kind:     "Role",
		Name:     r.roleAndBindingName(approval.Name),
	}

	// Set users in role bindings
	for _, u := range approval.Spec.Users {
		// Default is user
		apiGroup := rbac.GroupName
		kind := rbac.UserKind
		name := u.Name
		ns := approval.Namespace

		// Check if it's ServiceAccount
		if strings.HasPrefix(u.Name, serviceAccountPrefix) {
			apiGroup = ""
			kind = rbac.ServiceAccountKind

			token := strings.Split(strings.TrimPrefix(u.Name, serviceAccountPrefix), ":")
			if len(token) != 2 {
				continue
			}
			ns = token[0]
			name = token[1]
		}

		// Check if exists in RoleBinding
		subjectExist := false
		for _, s := range binding.Subjects {
			if s.APIGroup == apiGroup && s.Kind == kind && s.Namespace == ns && s.Name == name {
				subjectExist = true
				break
			}
		}

		if subjectExist {
			continue
		}

		binding.Subjects = append(binding.Subjects, rbac.Subject{
			APIGroup:  apiGroup,
			Kind:      kind,
			Name:      name,
			Namespace: ns,
		})
	}

	return utils.CreateOrPatchObject(binding, original, approval, r.Client, r.Scheme)
}

func (r *ApprovalReconciler) labelsForRoleAndBinding(approval *cicdv1.Approval) map[string]string {
	result := map[string]string{
		cicdv1.JobLabelPrefix + "approval":       approval.Name,
		cicdv1.JobLabelPrefix + "integrationJob": approval.Spec.IntegrationJob,
	}

	return result
}

func (r *ApprovalReconciler) annotationsForRoleAndBinding(approval *cicdv1.Approval) map[string]string {
	result := map[string]string{}
	if approval.Spec.Sender != nil {
		result[cicdv1.JobLabelPrefix+"sender"] = approval.Spec.Sender.Name
	}
	return result
}

func (r *ApprovalReconciler) generateMail(instance *cicdv1.Approval, titleKey, contentKey string) (string, string, error) {
	templatePath := os.Getenv(configMapPathEnv)
	if templatePath == "" {
		templatePath = configMapPathDefault
	}

	// Get template
	titleTemplate, err := template.ParseFiles(path.Join(templatePath, titleKey))
	if err != nil {
		return "", "", err
	}

	contentTemplate, err := template.ParseFiles(path.Join(templatePath, contentKey))
	if err != nil {
		return "", "", err
	}

	// Generate mail title, content
	title := &bytes.Buffer{}
	if err := titleTemplate.Execute(title, instance); err != nil {
		return "", "", err
	}
	content := &bytes.Buffer{}
	if err := contentTemplate.Execute(content, instance); err != nil {
		return "", "", err
	}

	return title.String(), content.String(), nil
}

func (r *ApprovalReconciler) sendMail(users []string, title, content string, cond *metav1.Condition) {
	cond.Status = metav1.ConditionFalse

	if !configs.EnableMail {
		cond.Reason = "EmailDisabled"
		cond.Message = "Email feature is disabled"
		return
	}

	// Send mail
	if err := r.MailSender.Send(users, title, content, true); err != nil {
		r.Log.Error(err, "")
		cond.Reason = "ErrorSendingMail"
		cond.Message = err.Error()
		return
	}

	cond.Status = metav1.ConditionTrue
	cond.Reason = "EmailSent"
	cond.Message = "Email is sent"
}

func (r *ApprovalReconciler) setError(instance *cicdv1.Approval, message string) {
	instance.Status.Result = cicdv1.ApprovalResultError
	instance.Status.Reason = message
}

// SetupWithManager sets ApprovalReconciler to the manager
func (r *ApprovalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cicdv1.Approval{}).
		Complete(r)
}
