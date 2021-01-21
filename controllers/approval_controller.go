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
	"bytes"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-lib/status"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/mail"
	"html/template"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
)

// ApprovalReconciler reconciles a Approval object
type ApprovalReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	ApiGroup           = "cicdapi.tmax.io"
	ApprovalKind       = "approvals"
	ApprovalSubApprove = "approve"
	ApprovalSubReject  = "reject"

	ServiceAccountPrefix = "system:serviceaccount:"
)

// +kubebuilder:rbac:groups=cicd.tmax.io,resources=approvals,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cicd.tmax.io,resources=approvals/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cicdapi.tmax.io,resources=approvals/approve;approvals/reject,verbs=update
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete

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
	original := instance.DeepCopy()

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
		p := client.MergeFrom(original)
		if err := r.Client.Status().Patch(ctx, instance, p); err != nil {
			log.Error(err, "")
		}
	}()

	// Default conditions
	if instance.Status.Result == "" {
		instance.Status.Result = cicdv1.ApprovalResultWaiting
	}

	// Create Role/RoleBinding if not exists
	if err := r.createOrUpdateRole(instance); err != nil {
		log.Error(err, "updating roles")
		r.patchStatus(instance, original, err.Error())
		return ctrl.Result{}, nil
	}
	if err := r.createOrUpdateRoleBinding(instance); err != nil {
		log.Error(err, "updating rolebindings")
		r.patchStatus(instance, original, err.Error())
		return ctrl.Result{}, nil
	}

	// Set SentRequestMail
	if !instance.Spec.SkipSendMail && (sentReqMailCond == nil || sentReqMailCond.Status == corev1.ConditionUnknown) {
		title, content, err := r.generateMail(instance, &configs.ApprovalRequestMailTitle, &configs.ApprovalRequestMailContent)
		if err != nil {
			log.Error(err, "")
			r.patchStatus(instance, original, err.Error())
			return ctrl.Result{}, nil
		}
		r.sendMail(instance.Spec.Users, title, content, sentReqMailCond)
	}

	// Set SentResultMail - only if is decided
	if !instance.Spec.SkipSendMail && (instance.Status.Result == cicdv1.ApprovalResultApproved || instance.Status.Result == cicdv1.ApprovalResultRejected) {
		if sentResMailCond == nil || sentResMailCond.Status == corev1.ConditionUnknown {
			title, content, err := r.generateMail(instance, &configs.ApprovalResultMailTitle, &configs.ApprovalResultMailContent)
			if err != nil {
				log.Error(err, "")
				r.patchStatus(instance, original, err.Error())
				return ctrl.Result{}, nil
			}
			r.sendMail(instance.Spec.Users, title, content, sentResMailCond)
		}
	}

	return ctrl.Result{}, nil
}

func roleAndBindingName(approvalName string) string {
	return "cicd-approval-" + approvalName
}

func (r *ApprovalReconciler) createOrUpdateRole(approval *cicdv1.Approval) error {
	notExist := false
	role := &rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleAndBindingName(approval.Name),
			Namespace: approval.Namespace,
			Labels:    labelsForRoleAndBinding(approval),
		},
	}
	if err := r.Client.Get(context.Background(), types.NamespacedName{Name: role.Name, Namespace: role.Namespace}, role); err != nil {
		if errors.IsNotFound(err) {
			notExist = true
		} else {
			return err
		}
	}
	original := role.DeepCopy()

	// Set rules
	role.Rules = []rbac.PolicyRule{{
		APIGroups:     []string{ApiGroup},
		Resources:     []string{fmt.Sprintf("%s/%s", ApprovalKind, ApprovalSubApprove), fmt.Sprintf("%s/%s", ApprovalKind, ApprovalSubReject)},
		ResourceNames: []string{approval.Name},
		Verbs:         []string{"update"},
	}}

	if err := controllerutil.SetOwnerReference(approval, role, r.Scheme); err != nil {
		return err
	}

	if notExist {
		// Create
		if err := r.Client.Create(context.Background(), role); err != nil {
			return err
		}
	} else {
		// Update
		p := client.MergeFrom(original)
		if err := r.Client.Patch(context.Background(), role, p); err != nil {
			return err
		}
	}

	return nil
}

func (r *ApprovalReconciler) createOrUpdateRoleBinding(approval *cicdv1.Approval) error {
	notExist := false
	binding := &rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleAndBindingName(approval.Name),
			Namespace: approval.Namespace,
			Labels:    labelsForRoleAndBinding(approval),
		},
	}
	if err := r.Client.Get(context.Background(), types.NamespacedName{Name: binding.Name, Namespace: binding.Namespace}, binding); err != nil {
		if errors.IsNotFound(err) {
			notExist = true
		} else {
			return err
		}
	}
	original := binding.DeepCopy()

	// Set bindings
	binding.RoleRef = rbac.RoleRef{
		APIGroup: rbac.GroupName,
		Kind:     "Role",
		Name:     roleAndBindingName(approval.Name),
	}

	for _, u := range approval.Spec.Users {
		token := strings.Split(u, "=")
		user := token[0]

		// Default is user
		apiGroup := rbac.GroupName
		kind := rbac.UserKind
		name := user
		ns := approval.Namespace

		// Check if it's ServiceAccount
		if strings.HasPrefix(user, ServiceAccountPrefix) {
			apiGroup = ""
			kind = rbac.ServiceAccountKind

			token := strings.Split(strings.TrimPrefix(user, ServiceAccountPrefix), ":")
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

	if err := controllerutil.SetOwnerReference(approval, binding, r.Scheme); err != nil {
		return err
	}

	if notExist {
		// Create
		if err := r.Client.Create(context.Background(), binding); err != nil {
			return err
		}
	} else {
		// Update
		p := client.MergeFrom(original)
		if err := r.Client.Patch(context.Background(), binding, p); err != nil {
			return err
		}
	}

	return nil
}

func labelsForRoleAndBinding(approval *cicdv1.Approval) map[string]string {
	return map[string]string{
		cicdv1.JobLabelPrefix + "approval":       approval.Name,
		cicdv1.JobLabelPrefix + "integrationJob": approval.Spec.IntegrationJob,
		cicdv1.JobLabelPrefix + "sender":         approval.Spec.Sender,
	}
}

func (r *ApprovalReconciler) generateMail(instance *cicdv1.Approval, titleTemplateCfg, contentTemplateCfg *string) (string, string, error) {
	// Get template
	var err error
	titleTemplate := template.New("")
	titleTemplate, err = titleTemplate.Parse(*titleTemplateCfg)
	if err != nil {
		return "", "", err
	}

	contentTemplate := template.New("")
	contentTemplate, err = contentTemplate.Parse(*contentTemplateCfg)
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

func (r *ApprovalReconciler) sendMail(users []string, title, content string, cond *status.Condition) {
	sentMail := false
	cond.Reason = "EmailDisabled"
	if configs.EnableMail {
		tos := utils.ParseEmailFromUsers(users)
		if err := mail.Send(tos, title, content, true, r.Client); err != nil {
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

func (r *ApprovalReconciler) patchStatus(instance *cicdv1.Approval, original *cicdv1.Approval, message string) {
	instance.Status.Result = cicdv1.ApprovalResultError
	instance.Status.Reason = message
	p := client.MergeFrom(original)
	if err := r.Client.Status().Patch(context.Background(), instance, p); err != nil {
		r.Log.Error(err, "")
	}
}

func (r *ApprovalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cicdv1.Approval{}).
		Owns(&rbac.Role{}).
		Owns(&rbac.RoleBinding{}).
		Complete(r)
}
