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
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/test"
	"github.com/tmax-cloud/cicd-operator/pkg/notification/mail"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestApprovalReconciler_Reconcile(t *testing.T) {
	s := runtime.NewScheme()
	require.NoError(t, cicdv1.AddToScheme(s))
	require.NoError(t, rbac.AddToScheme(s))

	sNoCicd := runtime.NewScheme()
	require.NoError(t, rbac.AddToScheme(sNoCicd))

	sNoRole := runtime.NewScheme()
	require.NoError(t, cicdv1.AddToScheme(sNoRole))
	sNoRole.AddKnownTypes(rbac.SchemeGroupVersion, &rbac.RoleBinding{})
	metav1.AddToGroupVersion(sNoRole, rbac.SchemeGroupVersion)

	sNoRoleBinding := runtime.NewScheme()
	require.NoError(t, cicdv1.AddToScheme(sNoRoleBinding))
	sNoRoleBinding.AddKnownTypes(rbac.SchemeGroupVersion, &rbac.Role{})
	metav1.AddToGroupVersion(sNoRoleBinding, rbac.SchemeGroupVersion)

	tc := map[string]struct {
		approval *cicdv1.Approval
		scheme   *runtime.Scheme
		template []byte

		errorOccurs         bool
		errorMessage        string
		expectedResult      cicdv1.ApprovalResult
		expectedReason      string
		expectedReqCond     metav1.Condition
		expectedResCond     metav1.Condition
		expectedRole        *rbac.Role
		expectedRoleBinding *rbac.RoleBinding
		expectedMails       []mail.FakeMailEntity
	}{
		"normal": {
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-approval",
					Namespace: "test-ns",
				},
				Spec: cicdv1.ApprovalSpec{
					Sender: &cicdv1.ApprovalUser{Name: "user", Email: "user@tmax.co.kr"},
					Users: []cicdv1.ApprovalUser{
						{Name: "admin", Email: "admin@tmax.co.kr"},
					},
				},
				Status: cicdv1.ApprovalStatus{
					Result: cicdv1.ApprovalResultAwaiting,
				},
			},
			scheme:          s,
			template:        []byte("HI - {{.Name}}"),
			expectedResult:  cicdv1.ApprovalResultAwaiting,
			expectedReason:  "",
			expectedReqCond: metav1.Condition{Status: metav1.ConditionTrue, Reason: "EmailSent", Message: "Email is sent"},
			expectedResCond: metav1.Condition{Status: metav1.ConditionUnknown, Reason: "NotProcessed", Message: "Result email is not processed yet"},
			expectedMails: []mail.FakeMailEntity{
				{Receivers: []string{"admin@tmax.co.kr"}, Title: "HI - test-approval", Content: "HI - test-approval", IsHTML: true},
			},
			expectedRole: &rbac.Role{
				Rules: []rbac.PolicyRule{
					{
						APIGroups:     []string{"cicdapi.tmax.io"},
						Resources:     []string{"approvals/approve", "approvals/reject"},
						ResourceNames: []string{"test-approval"},
						Verbs:         []string{"update"},
					},
				},
			},
			expectedRoleBinding: &rbac.RoleBinding{
				RoleRef: rbac.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "Role", Name: "cicd-approval-test-approval"},
				Subjects: []rbac.Subject{
					{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: "admin", Namespace: "test-ns"},
				},
			},
		},
		"getErr": {
			scheme:       sNoCicd,
			errorOccurs:  true,
			errorMessage: "no kind is registered for the type v1.Approval in scheme \"pkg/runtime/scheme.go:100\"",
		},
		"notFound": {
			scheme: s,
		},
		"default": {
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-approval",
					Namespace: "test-ns",
				},
				Spec: cicdv1.ApprovalSpec{
					Sender: &cicdv1.ApprovalUser{Name: "user", Email: "user@tmax.co.kr"},
					Users: []cicdv1.ApprovalUser{
						{Name: "admin", Email: "admin@tmax.co.kr"},
					},
				},
			},
			scheme:          s,
			template:        []byte("HI - {{.Name}}"),
			expectedResult:  cicdv1.ApprovalResultAwaiting,
			expectedReason:  "",
			expectedReqCond: metav1.Condition{Status: metav1.ConditionUnknown, Reason: "NotProcessed", Message: "Request email is not processed yet"},
			expectedResCond: metav1.Condition{Status: metav1.ConditionUnknown, Reason: "NotProcessed", Message: "Result email is not processed yet"},
		},
		"bumpV0.5.0": {
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-approval",
					Namespace: "test-ns",
				},
				Spec: cicdv1.ApprovalSpec{
					Sender: &cicdv1.ApprovalUser{Name: "user", Email: "user@tmax.co.kr"},
					Users: []cicdv1.ApprovalUser{
						{Name: "admin", Email: "admin@tmax.co.kr"},
					},
				},
				Status: cicdv1.ApprovalStatus{
					Conditions: []metav1.Condition{
						{Type: cicdv1.ApprovalConditionSentRequestMail, Status: metav1.ConditionUnknown},
						{Type: cicdv1.ApprovalConditionSentResultMail, Status: metav1.ConditionUnknown},
					},
				},
			},
			scheme:          s,
			template:        []byte("HI - {{.Name}}"),
			expectedResult:  cicdv1.ApprovalResultAwaiting,
			expectedReason:  "",
			expectedReqCond: metav1.Condition{Status: metav1.ConditionUnknown, Reason: "Unknown", Message: "Unknown"},
			expectedResCond: metav1.Condition{Status: metav1.ConditionUnknown, Reason: "Unknown", Message: "Unknown"},
		},
		"createRoleErr": {
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-approval",
					Namespace: "test-ns",
				},
				Spec: cicdv1.ApprovalSpec{
					Sender: &cicdv1.ApprovalUser{Name: "user", Email: "user@tmax.co.kr"},
					Users: []cicdv1.ApprovalUser{
						{Name: "admin", Email: "admin@tmax.co.kr"},
					},
				},
				Status: cicdv1.ApprovalStatus{
					Result: cicdv1.ApprovalResultAwaiting,
				},
			},
			scheme:          sNoRole,
			template:        []byte("HI - {{.Name}}"),
			expectedResult:  cicdv1.ApprovalResultError,
			expectedReason:  "no kind is registered for the type v1.Role in scheme \"pkg/runtime/scheme.go:100\"",
			expectedReqCond: metav1.Condition{Status: metav1.ConditionUnknown, Reason: "NotProcessed", Message: "Request email is not processed yet"},
			expectedResCond: metav1.Condition{Status: metav1.ConditionUnknown, Reason: "NotProcessed", Message: "Result email is not processed yet"},
		},
		"createRbErr": {
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-approval",
					Namespace: "test-ns",
				},
				Spec: cicdv1.ApprovalSpec{
					Sender: &cicdv1.ApprovalUser{Name: "user", Email: "user@tmax.co.kr"},
					Users: []cicdv1.ApprovalUser{
						{Name: "admin", Email: "admin@tmax.co.kr"},
					},
				},
				Status: cicdv1.ApprovalStatus{
					Result: cicdv1.ApprovalResultAwaiting,
				},
			},
			scheme:          sNoRoleBinding,
			template:        []byte("HI - {{.Name}}"),
			expectedResult:  cicdv1.ApprovalResultError,
			expectedReason:  "no kind is registered for the type v1.RoleBinding in scheme \"pkg/runtime/scheme.go:100\"",
			expectedReqCond: metav1.Condition{Status: metav1.ConditionUnknown, Reason: "NotProcessed", Message: "Request email is not processed yet"},
			expectedResCond: metav1.Condition{Status: metav1.ConditionUnknown, Reason: "NotProcessed", Message: "Result email is not processed yet"},
			expectedRole: &rbac.Role{
				Rules: []rbac.PolicyRule{
					{
						APIGroups:     []string{"cicdapi.tmax.io"},
						Resources:     []string{"approvals/approve", "approvals/reject"},
						ResourceNames: []string{"test-approval"},
						Verbs:         []string{"update"},
					},
				},
			},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			configs.EnableMail = true
			templatePath := path.Join(os.TempDir(), "cicd-test-process-mail-template")
			require.NoError(t, os.Setenv(configMapPathEnv, templatePath))
			defer func() {
				_ = os.Unsetenv(configMapPathEnv)
			}()
			require.NoError(t, os.MkdirAll(templatePath, os.ModePerm))
			defer func() {
				_ = os.RemoveAll(templatePath)
			}()
			require.NoError(t, ioutil.WriteFile(path.Join(templatePath, configMapKeyRequestTitle), c.template, os.ModePerm))
			require.NoError(t, ioutil.WriteFile(path.Join(templatePath, configMapKeyRequestContent), c.template, os.ModePerm))
			require.NoError(t, ioutil.WriteFile(path.Join(templatePath, configMapKeyResultTitle), c.template, os.ModePerm))
			require.NoError(t, ioutil.WriteFile(path.Join(templatePath, configMapKeyResultContent), c.template, os.ModePerm))

			sender := mail.NewFakeSender()
			fakeCli := fake.NewClientBuilder().WithScheme(c.scheme).Build()
			if c.approval != nil {
				require.NoError(t, fakeCli.Create(context.Background(), c.approval))
			}
			reconciler := &ApprovalReconciler{Client: fakeCli, MailSender: sender, Log: &test.FakeLogger{}, Scheme: c.scheme}
			_, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: "test-approval", Namespace: "test-ns"}})
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)

				if c.approval == nil {
					return
				}
				result := &cicdv1.Approval{}
				require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: c.approval.Name, Namespace: c.approval.Namespace}, result))

				require.Equal(t, c.expectedResult, result.Status.Result)
				require.Equal(t, c.expectedReason, result.Status.Reason)
				require.Equal(t, c.expectedReqCond.Status, meta.FindStatusCondition(result.Status.Conditions, "SentRequestMail").Status)
				require.Equal(t, c.expectedReqCond.Reason, meta.FindStatusCondition(result.Status.Conditions, "SentRequestMail").Reason)
				require.Equal(t, c.expectedReqCond.Message, meta.FindStatusCondition(result.Status.Conditions, "SentRequestMail").Message)
				require.Equal(t, c.expectedResCond.Status, meta.FindStatusCondition(result.Status.Conditions, "SentResultMail").Status)
				require.Equal(t, c.expectedResCond.Reason, meta.FindStatusCondition(result.Status.Conditions, "SentResultMail").Reason)
				require.Equal(t, c.expectedResCond.Message, meta.FindStatusCondition(result.Status.Conditions, "SentResultMail").Message)
				require.Equal(t, c.expectedReason, result.Status.Reason)
				require.Equal(t, c.expectedMails, sender.Mails)
				if c.expectedRole != nil {
					role := &rbac.Role{}
					require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: "cicd-approval-" + c.approval.Name, Namespace: c.approval.Namespace}, role))
					require.Equal(t, c.expectedRole.Rules, role.Rules)
				}
				if c.expectedRoleBinding != nil {
					rb := &rbac.RoleBinding{}
					require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: "cicd-approval-" + c.approval.Name, Namespace: c.approval.Namespace}, rb))
					require.Equal(t, c.expectedRoleBinding.RoleRef, rb.RoleRef)
					require.Equal(t, c.expectedRoleBinding.Subjects, rb.Subjects)
				}
			}
		})
	}
}

func TestApprovalReconciler_bumpV050(t *testing.T) {
	reconciler := &ApprovalReconciler{}

	ic := &cicdv1.Approval{
		Status: cicdv1.ApprovalStatus{
			Conditions: []metav1.Condition{
				{Type: cicdv1.ApprovalConditionSentRequestMail, Status: metav1.ConditionTrue},
				{Type: cicdv1.ApprovalConditionSentResultMail, Status: metav1.ConditionTrue},
			},
		},
	}
	reconciler.bumpV050(ic)

	require.Equal(t, "Sent", ic.Status.Conditions[0].Reason)
	require.Equal(t, "Sent", ic.Status.Conditions[0].Message)

	require.Equal(t, "Sent", ic.Status.Conditions[1].Reason)
	require.Equal(t, "Sent", ic.Status.Conditions[1].Message)
}

func TestApprovalReconciler_processMail(t *testing.T) {
	tc := map[string]struct {
		approval        *cicdv1.Approval
		requestTemplate []byte
		resultTemplate  []byte

		expectedReqCond metav1.Condition
		expectedResCond metav1.Condition
		expectedMails   []mail.FakeMailEntity
	}{
		"sendAll": {
			approval: &cicdv1.Approval{
				Spec: cicdv1.ApprovalSpec{
					Sender: &cicdv1.ApprovalUser{Name: "admin", Email: "admin@tmax.co.kr"},
					Users:  []cicdv1.ApprovalUser{{Name: "test", Email: "test@tmax.co.kr"}},
				},
				Status: cicdv1.ApprovalStatus{
					Result: cicdv1.ApprovalResultApproved,
					Conditions: []metav1.Condition{
						{Type: cicdv1.ApprovalConditionSentRequestMail, Status: metav1.ConditionUnknown},
						{Type: cicdv1.ApprovalConditionSentResultMail, Status: metav1.ConditionUnknown},
					},
				},
			},
			requestTemplate: []byte("{{.Spec.Sender.Name}} - request"),
			resultTemplate:  []byte("{{.Spec.Sender.Name}} - result"),

			expectedReqCond: metav1.Condition{Type: "SentRequestMail", Status: metav1.ConditionTrue, Reason: "EmailSent", Message: "Email is sent"},
			expectedResCond: metav1.Condition{Type: "SentResultMail", Status: metav1.ConditionTrue, Reason: "EmailSent", Message: "Email is sent"},
			expectedMails: []mail.FakeMailEntity{
				{Receivers: []string{"test@tmax.co.kr"}, Title: "admin - request", Content: "admin - request", IsHTML: true},
				{Receivers: []string{"admin@tmax.co.kr"}, Title: "admin - result", Content: "admin - result", IsHTML: true},
			},
		},
		"skipSendMail": {
			approval: &cicdv1.Approval{
				Spec: cicdv1.ApprovalSpec{
					SkipSendMail: true,
					Sender:       &cicdv1.ApprovalUser{Name: "admin", Email: "admin@tmax.co.kr"},
					Users:        []cicdv1.ApprovalUser{{Name: "test", Email: "test@tmax.co.kr"}},
				},
				Status: cicdv1.ApprovalStatus{
					Result: cicdv1.ApprovalResultApproved,
					Conditions: []metav1.Condition{
						{Type: cicdv1.ApprovalConditionSentRequestMail, Status: metav1.ConditionUnknown},
						{Type: cicdv1.ApprovalConditionSentResultMail, Status: metav1.ConditionUnknown},
					},
				},
			},
			requestTemplate: []byte("{{.Spec.Sender.Name}} - request"),
			resultTemplate:  []byte("{{.Spec.Sender.Name}} - result"),

			expectedReqCond: metav1.Condition{Type: "SentRequestMail", Status: metav1.ConditionUnknown},
			expectedResCond: metav1.Condition{Type: "SentResultMail", Status: metav1.ConditionUnknown},
		},
		"reqGenErr": {
			approval: &cicdv1.Approval{
				Spec: cicdv1.ApprovalSpec{
					Sender: &cicdv1.ApprovalUser{Name: "admin", Email: "admin@tmax.co.kr"},
					Users:  []cicdv1.ApprovalUser{{Name: "test", Email: "test@tmax.co.kr"}},
				},
				Status: cicdv1.ApprovalStatus{
					Result: cicdv1.ApprovalResultApproved,
					Conditions: []metav1.Condition{
						{Type: cicdv1.ApprovalConditionSentRequestMail, Status: metav1.ConditionUnknown},
						{Type: cicdv1.ApprovalConditionSentResultMail, Status: metav1.ConditionUnknown},
					},
				},
			},
			requestTemplate: []byte("{{....Spec.Sender.Name}} - request"),
			resultTemplate:  []byte("{{.Spec.Sender.Name}} - result"),

			expectedReqCond: metav1.Condition{Type: "SentRequestMail", Status: metav1.ConditionFalse, Reason: "EmailGenerateError", Message: "template: request-title:1: unexpected <.> in operand"},
			expectedResCond: metav1.Condition{Type: "SentResultMail", Status: metav1.ConditionTrue, Reason: "EmailSent", Message: "Email is sent"},
			expectedMails: []mail.FakeMailEntity{
				{Receivers: []string{"admin@tmax.co.kr"}, Title: "admin - result", Content: "admin - result", IsHTML: true},
			},
		},
		"approvalNotDecided": {
			approval: &cicdv1.Approval{
				Spec: cicdv1.ApprovalSpec{
					Sender: &cicdv1.ApprovalUser{Name: "admin", Email: "admin@tmax.co.kr"},
					Users:  []cicdv1.ApprovalUser{{Name: "test", Email: "test@tmax.co.kr"}},
				},
				Status: cicdv1.ApprovalStatus{
					Conditions: []metav1.Condition{
						{Type: cicdv1.ApprovalConditionSentRequestMail, Status: metav1.ConditionUnknown},
						{Type: cicdv1.ApprovalConditionSentResultMail, Status: metav1.ConditionUnknown},
					},
				},
			},
			requestTemplate: []byte("{{.Spec.Sender.Name}} - request"),
			resultTemplate:  []byte("{{.Spec.Sender.Name}} - result"),

			expectedReqCond: metav1.Condition{Type: "SentRequestMail", Status: metav1.ConditionTrue, Reason: "EmailSent", Message: "Email is sent"},
			expectedResCond: metav1.Condition{Type: "SentResultMail", Status: metav1.ConditionUnknown},
			expectedMails: []mail.FakeMailEntity{
				{Receivers: []string{"test@tmax.co.kr"}, Title: "admin - request", Content: "admin - request", IsHTML: true},
			},
		},
		"approvalSenderNil": {
			approval: &cicdv1.Approval{
				Spec: cicdv1.ApprovalSpec{
					Users: []cicdv1.ApprovalUser{{Name: "test", Email: "test@tmax.co.kr"}},
				},
				Status: cicdv1.ApprovalStatus{
					Result: cicdv1.ApprovalResultApproved,
					Conditions: []metav1.Condition{
						{Type: cicdv1.ApprovalConditionSentRequestMail, Status: metav1.ConditionUnknown},
						{Type: cicdv1.ApprovalConditionSentResultMail, Status: metav1.ConditionUnknown},
					},
				},
			},
			requestTemplate: []byte("{{.Name}} - request"),
			resultTemplate:  []byte("{{.Name}} - result"),

			expectedReqCond: metav1.Condition{Type: "SentRequestMail", Status: metav1.ConditionTrue, Reason: "EmailSent", Message: "Email is sent"},
			expectedResCond: metav1.Condition{Type: "SentResultMail", Status: metav1.ConditionUnknown},
			expectedMails: []mail.FakeMailEntity{
				{Receivers: []string{"test@tmax.co.kr"}, Title: " - request", Content: " - request", IsHTML: true},
			},
		},
		"resGenErr": {
			approval: &cicdv1.Approval{
				Spec: cicdv1.ApprovalSpec{
					Sender: &cicdv1.ApprovalUser{Name: "admin", Email: "admin@tmax.co.kr"},
					Users:  []cicdv1.ApprovalUser{{Name: "test", Email: "test@tmax.co.kr"}},
				},
				Status: cicdv1.ApprovalStatus{
					Result: cicdv1.ApprovalResultApproved,
					Conditions: []metav1.Condition{
						{Type: cicdv1.ApprovalConditionSentRequestMail, Status: metav1.ConditionUnknown},
						{Type: cicdv1.ApprovalConditionSentResultMail, Status: metav1.ConditionUnknown},
					},
				},
			},
			requestTemplate: []byte("{{.Spec.Sender.Name}} - request"),
			resultTemplate:  []byte("{{....Spec.Sender.Name}} - result"),

			expectedReqCond: metav1.Condition{Type: "SentRequestMail", Status: metav1.ConditionTrue, Reason: "EmailSent", Message: "Email is sent"},
			expectedResCond: metav1.Condition{Type: "SentResultMail", Status: metav1.ConditionFalse, Reason: "EmailGenerateError", Message: "template: result-title:1: unexpected <.> in operand"},
			expectedMails: []mail.FakeMailEntity{
				{Receivers: []string{"test@tmax.co.kr"}, Title: "admin - request", Content: "admin - request", IsHTML: true},
			},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			configs.EnableMail = true
			templatePath := path.Join(os.TempDir(), "cicd-test-process-mail-template")
			require.NoError(t, os.Setenv(configMapPathEnv, templatePath))
			defer func() {
				_ = os.Unsetenv(configMapPathEnv)
			}()
			require.NoError(t, os.MkdirAll(templatePath, os.ModePerm))
			defer func() {
				_ = os.RemoveAll(templatePath)
			}()
			require.NoError(t, ioutil.WriteFile(path.Join(templatePath, configMapKeyRequestTitle), c.requestTemplate, os.ModePerm))
			require.NoError(t, ioutil.WriteFile(path.Join(templatePath, configMapKeyRequestContent), c.requestTemplate, os.ModePerm))
			require.NoError(t, ioutil.WriteFile(path.Join(templatePath, configMapKeyResultTitle), c.resultTemplate, os.ModePerm))
			require.NoError(t, ioutil.WriteFile(path.Join(templatePath, configMapKeyResultContent), c.resultTemplate, os.ModePerm))

			sender := mail.NewFakeSender()
			reconciler := &ApprovalReconciler{MailSender: sender, Log: &test.FakeLogger{}}

			reconciler.processMail(c.approval)
			require.Equal(t, c.expectedReqCond, *meta.FindStatusCondition(c.approval.Status.Conditions, cicdv1.ApprovalConditionSentRequestMail))
			require.Equal(t, c.expectedResCond, *meta.FindStatusCondition(c.approval.Status.Conditions, cicdv1.ApprovalConditionSentResultMail))
			require.Equal(t, c.expectedMails, sender.Mails)
		})
	}
}

func TestApprovalReconciler_roleAndBindingName(t *testing.T) {
	reconciler := &ApprovalReconciler{}
	name := reconciler.roleAndBindingName("test-approval")
	require.Equal(t, "cicd-approval-test-approval", name)
}

func TestApprovalReconciler_createOrUpdateRole(t *testing.T) {
	s := runtime.NewScheme()
	require.NoError(t, cicdv1.AddToScheme(s))
	require.NoError(t, rbac.AddToScheme(s))

	sNoRbac := runtime.NewScheme()
	require.NoError(t, cicdv1.AddToScheme(sNoRbac))

	tc := map[string]struct {
		scheme   *runtime.Scheme
		approval *cicdv1.Approval

		errorOccurs         bool
		errorMessage        string
		expectedLabels      map[string]string
		expectedAnnotations map[string]string
		expectedRules       []rbac.PolicyRule
	}{
		"normal": {
			scheme: s,
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-approval",
					Namespace: "default",
				},
				Spec: cicdv1.ApprovalSpec{
					IntegrationJob: "test-ij",
					Sender: &cicdv1.ApprovalUser{
						Name:  "admin@tmax.co.kr",
						Email: "admin@tmax.co.kr",
					},
					Users: []cicdv1.ApprovalUser{
						{Name: "user1", Email: "user1@tmax.co.kr"},
						{Name: "system:serviceaccount:default:user1"},
						{Name: "system:serviceaccount:default"},
					},
				},
			},
			expectedLabels: map[string]string{
				"cicd.tmax.io/approval":       "test-approval",
				"cicd.tmax.io/integrationJob": "test-ij",
			},
			expectedAnnotations: map[string]string{
				"cicd.tmax.io/sender": "admin@tmax.co.kr",
			},
			expectedRules: []rbac.PolicyRule{
				{
					APIGroups:     []string{"cicdapi.tmax.io"},
					Resources:     []string{"approvals/approve", "approvals/reject"},
					ResourceNames: []string{"test-approval"},
					Verbs:         []string{"update"},
				},
			},
		},
		"getErr": {
			scheme: sNoRbac,
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-approval",
					Namespace: "default",
				},
				Spec: cicdv1.ApprovalSpec{
					IntegrationJob: "test-ij",
					Users: []cicdv1.ApprovalUser{
						{Name: "user1", Email: "user1@tmax.co.kr"},
						{Name: "system:serviceaccount:default:user1"},
						{Name: "system:serviceaccount:default"},
					},
				},
			},
			errorOccurs:  true,
			errorMessage: "no kind is registered for the type v1.Role in scheme \"pkg/runtime/scheme.go:100\"",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			fakeCli := fake.NewClientBuilder().WithScheme(c.scheme).Build()

			reconciler := &ApprovalReconciler{Client: fakeCli, Scheme: c.scheme}
			err := reconciler.createOrUpdateRole(c.approval)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)

				result := &rbac.Role{}
				require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: "cicd-approval-" + c.approval.Name, Namespace: c.approval.Namespace}, result))
				require.Equal(t, c.expectedLabels, result.Labels)
				require.Equal(t, c.expectedAnnotations, result.Annotations)
				require.Equal(t, c.expectedRules, result.Rules)
			}
		})
	}
}

func TestApprovalReconciler_createOrUpdateRoleBinding(t *testing.T) {
	s := runtime.NewScheme()
	require.NoError(t, cicdv1.AddToScheme(s))
	require.NoError(t, rbac.AddToScheme(s))

	sNoRbac := runtime.NewScheme()
	require.NoError(t, cicdv1.AddToScheme(sNoRbac))

	tc := map[string]struct {
		scheme   *runtime.Scheme
		approval *cicdv1.Approval
		rb       *rbac.RoleBinding

		errorOccurs         bool
		errorMessage        string
		expectedLabels      map[string]string
		expectedAnnotations map[string]string
		expectedSubject     []rbac.Subject
		expectedRoleRef     rbac.RoleRef
	}{
		"normal": {
			scheme: s,
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-approval",
					Namespace: "default",
				},
				Spec: cicdv1.ApprovalSpec{
					IntegrationJob: "test-ij",
					Sender: &cicdv1.ApprovalUser{
						Name:  "admin@tmax.co.kr",
						Email: "admin@tmax.co.kr",
					},
					Users: []cicdv1.ApprovalUser{
						{Name: "user1", Email: "user1@tmax.co.kr"},
						{Name: "system:serviceaccount:default:user1"},
						{Name: "system:serviceaccount:default"},
					},
				},
			},
			rb: &rbac.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cicd-approval-test-approval",
					Namespace: "default",
				},
				RoleRef: rbac.RoleRef{},
				Subjects: []rbac.Subject{
					{APIGroup: "", Kind: rbac.ServiceAccountKind, Namespace: "default", Name: "user1"},
				},
			},
			expectedLabels: map[string]string{
				"cicd.tmax.io/approval":       "test-approval",
				"cicd.tmax.io/integrationJob": "test-ij",
			},
			expectedAnnotations: map[string]string{
				"cicd.tmax.io/sender": "admin@tmax.co.kr",
			},
			expectedSubject: []rbac.Subject{
				{Kind: "ServiceAccount", Name: "user1", Namespace: "default"},
				{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: "user1", Namespace: "default"},
			},
			expectedRoleRef: rbac.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "Role", Name: "cicd-approval-test-approval"},
		},
		"getErr": {
			scheme: sNoRbac,
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-approval",
					Namespace: "default",
				},
				Spec: cicdv1.ApprovalSpec{
					IntegrationJob: "test-ij",
					Users: []cicdv1.ApprovalUser{
						{Name: "user1", Email: "user1@tmax.co.kr"},
						{Name: "system:serviceaccount:default:user1"},
						{Name: "system:serviceaccount:default"},
					},
				},
			},
			errorOccurs:  true,
			errorMessage: "no kind is registered for the type v1.RoleBinding in scheme \"pkg/runtime/scheme.go:100\"",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			fakeCli := fake.NewClientBuilder().WithScheme(c.scheme).Build()
			if c.rb != nil {
				require.NoError(t, fakeCli.Create(context.Background(), c.rb))
			}

			reconciler := &ApprovalReconciler{Client: fakeCli, Scheme: c.scheme}
			err := reconciler.createOrUpdateRoleBinding(c.approval)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)

				result := &rbac.RoleBinding{}
				require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: "cicd-approval-" + c.approval.Name, Namespace: c.approval.Namespace}, result))
				require.Equal(t, c.expectedLabels, result.Labels)
				require.Equal(t, c.expectedAnnotations, result.Annotations)
				require.Equal(t, c.expectedSubject, result.Subjects)
				require.Equal(t, c.expectedRoleRef, result.RoleRef)
			}
		})
	}
}

func TestApprovalReconciler_labelsForRoleAndBinding(t *testing.T) {
	approval := &cicdv1.Approval{
		ObjectMeta: metav1.ObjectMeta{Name: "test-approval"},
		Spec:       cicdv1.ApprovalSpec{IntegrationJob: "test-ij"},
	}

	reconciler := &ApprovalReconciler{}
	labels := reconciler.labelsForRoleAndBinding(approval)
	require.Equal(t, map[string]string{
		"cicd.tmax.io/approval":       "test-approval",
		"cicd.tmax.io/integrationJob": "test-ij",
	}, labels)
}

func TestApprovalReconciler_annotationsForRoleAndBinding(t *testing.T) {
	approval := &cicdv1.Approval{}

	tc := map[string]struct {
		sender *cicdv1.ApprovalUser

		expectedMap map[string]string
	}{
		"normal": {
			sender: &cicdv1.ApprovalUser{
				Name:  "admin",
				Email: "admin@tmax.co.kr",
			},
			expectedMap: map[string]string{
				"cicd.tmax.io/sender": "admin",
			},
		},
		"nilSender": {
			expectedMap: map[string]string{},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			approval.Spec.Sender = c.sender

			reconciler := &ApprovalReconciler{}
			result := reconciler.annotationsForRoleAndBinding(approval)
			require.Equal(t, c.expectedMap, result)
		})
	}
}

func TestApprovalReconciler_generateMail(t *testing.T) {
	require.NoError(t, cicdv1.AddToScheme(scheme.Scheme))
	approval := &cicdv1.Approval{
		Spec: cicdv1.ApprovalSpec{
			IntegrationJob: "test-ij",
			Message:        "ij message",
		},
	}

	titleKey := "title"
	contentKey := "content"

	tc := map[string]struct {
		path            string
		titleTemplate   []byte
		contentTemplate []byte

		errorOccurs     bool
		errorMessage    string
		expectedTitle   string
		expectedContent string
	}{
		"normal": {
			path:            path.Join(os.TempDir(), "cicd-test-approval-mail-template"),
			titleTemplate:   []byte("TITLE: {{.Spec.IntegrationJob}}"),
			contentTemplate: []byte("CONTENT: {{.Spec.Message}}"),

			expectedTitle:   "TITLE: test-ij",
			expectedContent: "CONTENT: ij message",
		},
		//"noPathEnv": {
		//	titleTemplate:   []byte("TITLE: {{.Spec.IntegrationJob}}"),
		//	contentTemplate: []byte("CONTENT: {{.Spec.Message}}"),
		//
		//	expectedTitle:   "TITLE: test-ij",
		//	expectedContent: "CONTENT: ij message",
		//},
		"titleTemplateParseErr": {
			path:            path.Join(os.TempDir(), "cicd-test-approval-mail-template"),
			titleTemplate:   []byte("TITLE: {{.....Spec.IntegrationJob}}"),
			contentTemplate: []byte("CONTENT: {{.Spec.Message}}"),
			errorOccurs:     true,
			errorMessage:    "template: title:1: unexpected <.> in operand",
		},
		"titleTemplateExecErr": {
			path:            path.Join(os.TempDir(), "cicd-test-approval-mail-template"),
			titleTemplate:   []byte("TITLE: {{.Spec.IntegrationJobbbb}}"),
			contentTemplate: []byte("CONTENT: {{.Spec.Message}}"),
			errorOccurs:     true,
			errorMessage:    "template: title:1:14: executing \"title\" at <.Spec.IntegrationJobbbb>: can't evaluate field IntegrationJobbbb in type v1.ApprovalSpec",
		},
		"contentTemplateParseErr": {
			path:            path.Join(os.TempDir(), "cicd-test-approval-mail-template"),
			titleTemplate:   []byte("TITLE: {{.Spec.IntegrationJob}}"),
			contentTemplate: []byte("CONTENT: {{...Spec.Message}}"),
			errorOccurs:     true,
			errorMessage:    "template: content:1: unexpected <.> in operand",
		},
		"contentTemplateExecErr": {
			path:            path.Join(os.TempDir(), "cicd-test-approval-mail-template"),
			titleTemplate:   []byte("TITLE: {{.Spec.IntegrationJob}}"),
			contentTemplate: []byte("CONTENT: {{.Spec.Messageeee}}"),
			errorOccurs:     true,
			errorMessage:    "template: content:1:16: executing \"content\" at <.Spec.Messageeee>: can't evaluate field Messageeee in type v1.ApprovalSpec",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			filePath := c.path
			if filePath == "" {
				filePath = configMapPathDefault
			} else {
				require.NoError(t, os.Setenv(configMapPathEnv, filePath))
				defer func() {
					_ = os.Unsetenv(configMapPathEnv)
				}()
			}

			require.NoError(t, os.MkdirAll(filePath, os.ModePerm))
			defer func() {
				_ = os.RemoveAll(filePath)
			}()
			require.NoError(t, ioutil.WriteFile(path.Join(filePath, titleKey), c.titleTemplate, os.ModePerm))
			require.NoError(t, ioutil.WriteFile(path.Join(filePath, contentKey), c.contentTemplate, os.ModePerm))

			fakeCli := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(approval).Build()

			reconciler := &ApprovalReconciler{Client: fakeCli}
			title, content, err := reconciler.generateMail(approval, titleKey, contentKey)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedTitle, title)
				require.Equal(t, c.expectedContent, content)
			}
		})
	}
}

func TestApprovalReconciler_sendMail(t *testing.T) {
	tc := map[string]struct {
		users    []string
		title    string
		content  string
		disabled bool

		condStatus  metav1.ConditionStatus
		condReason  string
		condMessage string
	}{
		"normal": {
			users:   []string{"admin@tmax.co.kr", "admin2@tmax.co.kr"},
			title:   "test-email",
			content: "TEST@@@@",

			condStatus:  metav1.ConditionTrue,
			condReason:  "EmailSent",
			condMessage: "Email is sent",
		},
		"errMail": {
			condStatus:  metav1.ConditionFalse,
			condReason:  "ErrorSendingMail",
			condMessage: "receivers list is empty",
		},
		"disabled": {
			disabled:    true,
			condStatus:  metav1.ConditionFalse,
			condReason:  "EmailDisabled",
			condMessage: "Email feature is disabled",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			configs.EnableMail = !c.disabled

			fakeLog := &test.FakeLogger{}
			sender := mail.NewFakeSender()
			reconciler := &ApprovalReconciler{Log: fakeLog, MailSender: sender}

			cond := &metav1.Condition{}
			reconciler.sendMail(c.users, c.title, c.content, cond)

			require.Equal(t, c.condStatus, cond.Status)
			require.Equal(t, c.condReason, cond.Reason)
			require.Equal(t, c.condMessage, cond.Message)

		})
	}
}

func TestApprovalReconciler_setError(t *testing.T) {
	utilruntime.Must(cicdv1.AddToScheme(scheme.Scheme))
	approval := &cicdv1.Approval{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-approval",
			Namespace: "test-ns",
		},
	}

	fakeLog := &test.FakeLogger{}
	fakeCli := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(approval).Build()

	reconciler := &ApprovalReconciler{Log: fakeLog, Client: fakeCli}
	reconciler.setError(approval, "just")

	require.Empty(t, fakeLog.Errors)

	require.Equal(t, cicdv1.ApprovalResultError, approval.Status.Result)
	require.Equal(t, "just", approval.Status.Reason)
}

func TestApprovalReconciler_SetupWithManager(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))

	mgr := &fakeManager{scheme: s}

	reconciler := &ApprovalReconciler{}
	require.NoError(t, reconciler.SetupWithManager(mgr))
}
