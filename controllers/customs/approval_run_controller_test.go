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
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"knative.dev/pkg/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	testApprovalRunName      = "test-approval"
	testApprovalRunNamespace = "test-ns"
)

type approvalHandlerTestCase struct {
	preFunc    func(t *testing.T, run *tektonv1alpha1.Run, handler *ApprovalRunHandler)
	verifyFunc func(t *testing.T, run *tektonv1alpha1.Run, approval *cicdv1.Approval)

	errorOccurs  bool
	errorMessage string
}

func TestApprovalRunHandler_Handle(t *testing.T) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}

	tc := map[string]approvalHandlerTestCase{
		"creation": {
			verifyFunc: func(t *testing.T, run *tektonv1alpha1.Run, approval *cicdv1.Approval) {
				// Verify Run status
				cond := run.Status.GetCondition(apis.ConditionSucceeded)
				require.NotNil(t, cond)
				require.True(t, cond.IsUnknown())
				require.Equal(t, "Awaiting", cond.Reason)
				require.Equal(t, "waiting for approval", cond.Message)

				// Verify created Approval
				require.Equal(t, run.Name, approval.Name)
				require.Equal(t, run.Namespace, approval.Namespace)
				require.False(t, approval.Spec.SkipSendMail)
				require.Equal(t, "none", approval.Spec.PipelineRun)
				require.Equal(t, "ij-sample", approval.Spec.IntegrationJob)
				require.Equal(t, "job-test", approval.Spec.JobName)
				require.Equal(t, "approval message!", approval.Spec.Message)
				require.NotNil(t, approval.Spec.Sender)
				require.Equal(t, "developer1", approval.Spec.Sender.Name)
				require.Equal(t, "dev@tmax.co.kr", approval.Spec.Sender.Email)
				require.Equal(t, "https://approval.ref", approval.Spec.Link)
				require.Len(t, approval.Spec.Users, 1)
				require.Equal(t, cicdv1.ApprovalUser{Name: "admin@tmax.co.kr", Email: "admin@tmax.co.kr"}, approval.Spec.Users[0])
			},
		},
		"creationNoMail": {
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run, handler *ApprovalRunHandler) {
				configs.EnableMail = false
			},
			verifyFunc: func(t *testing.T, run *tektonv1alpha1.Run, approval *cicdv1.Approval) {
				// Verify Run status
				cond := run.Status.GetCondition(apis.ConditionSucceeded)
				require.NotNil(t, cond)
				require.True(t, cond.IsUnknown())
				require.Equal(t, "Awaiting", cond.Reason)
				require.Equal(t, "waiting for approval", cond.Message)

				// Verify created Approval
				require.Equal(t, run.Name, approval.Name)
				require.Equal(t, run.Namespace, approval.Namespace)
				require.True(t, approval.Spec.SkipSendMail)
				require.Equal(t, "none", approval.Spec.PipelineRun)
				require.Equal(t, "ij-sample", approval.Spec.IntegrationJob)
				require.Equal(t, "job-test", approval.Spec.JobName)
				require.Equal(t, "approval message!", approval.Spec.Message)
				require.NotNil(t, approval.Spec.Sender)
				require.Equal(t, "developer1", approval.Spec.Sender.Name)
				require.Equal(t, "dev@tmax.co.kr", approval.Spec.Sender.Email)
				require.Equal(t, "https://approval.ref", approval.Spec.Link)
				require.Len(t, approval.Spec.Users, 1)
				require.Equal(t, cicdv1.ApprovalUser{Name: "admin@tmax.co.kr", Email: "admin@tmax.co.kr"}, approval.Spec.Users[0])
			},
		},
		"creationApproversCM": {
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run, handler *ApprovalRunHandler) {
				configMapName := "approver-cm-1"
				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: configMapName, Namespace: testApprovalRunNamespace},
					Data:       map[string]string{cicdv1.CustomTaskApprovalApproversConfigMapKey: "admin2@tmax.co.kr=admin2@tmax.co.kr,admin3@tmax.co.kr"},
				}
				require.NoError(t, handler.Client.Create(context.Background(), cm))
				run.Spec.Params[1].Value.StringVal = configMapName
			},
			verifyFunc: func(t *testing.T, run *tektonv1alpha1.Run, approval *cicdv1.Approval) {
				// Verify Run status
				cond := run.Status.GetCondition(apis.ConditionSucceeded)
				require.NotNil(t, cond)
				require.True(t, cond.IsUnknown())
				require.Equal(t, "Awaiting", cond.Reason)
				require.Equal(t, "waiting for approval", cond.Message)

				// Verify created Approval
				require.Equal(t, run.Name, approval.Name)
				require.Equal(t, run.Namespace, approval.Namespace)
				require.False(t, approval.Spec.SkipSendMail)
				require.Equal(t, "none", approval.Spec.PipelineRun)
				require.Equal(t, "ij-sample", approval.Spec.IntegrationJob)
				require.Equal(t, "job-test", approval.Spec.JobName)
				require.Equal(t, "approval message!", approval.Spec.Message)
				require.NotNil(t, approval.Spec.Sender)
				require.Equal(t, "developer1", approval.Spec.Sender.Name)
				require.Equal(t, "dev@tmax.co.kr", approval.Spec.Sender.Email)
				require.Equal(t, "https://approval.ref", approval.Spec.Link)
				require.Len(t, approval.Spec.Users, 3)
				require.Equal(t, cicdv1.ApprovalUser{Name: "admin@tmax.co.kr", Email: "admin@tmax.co.kr"}, approval.Spec.Users[0])
				require.Equal(t, cicdv1.ApprovalUser{Name: "admin2@tmax.co.kr", Email: "admin2@tmax.co.kr"}, approval.Spec.Users[1])
				require.Equal(t, cicdv1.ApprovalUser{Name: "admin3@tmax.co.kr"}, approval.Spec.Users[2])
			},
		},
		"createError": {
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run, handler *ApprovalRunHandler) {
				s := runtime.NewScheme()
				utilruntime.Must(corev1.AddToScheme(s))
				utilruntime.Must(cicdv1.AddToScheme(s))
				handler.Scheme = s
			},
			verifyFunc: func(t *testing.T, run *tektonv1alpha1.Run, approval *cicdv1.Approval) {
				// Verify Run status
				cond := run.Status.GetCondition(apis.ConditionSucceeded)
				require.NotNil(t, cond)
				require.True(t, cond.IsFalse())
				require.Equal(t, "CannotCreateApproval", cond.Reason)
				require.Equal(t, "no kind is registered for the type v1alpha1.Run in scheme \"pkg/runtime/scheme.go:100\"", cond.Message)
			},
		},
		"approvalApproved": {
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run, handler *ApprovalRunHandler) {
				tt, err := time.Parse("2006-01-02 15:04:05 -0700 MST", "2021-07-23 14:40:30 +0900 KST")
				require.NoError(t, err)

				require.NoError(t, createApprovalForRun(handler.Client, cicdv1.ApprovalStatus{
					Result:       cicdv1.ApprovalResultApproved,
					Approver:     "admin",
					Reason:       "good",
					DecisionTime: &metav1.Time{Time: tt},
				}))
			},
			verifyFunc: func(t *testing.T, run *tektonv1alpha1.Run, approval *cicdv1.Approval) {
				cond := run.Status.GetCondition(apis.ConditionSucceeded)
				require.NotNil(t, cond)
				require.True(t, cond.IsTrue())
				require.Equal(t, "Approved", cond.Reason)
				require.Equal(t, fmt.Sprintf("admin approved this approval, reason: good, decisionTime: %s", approval.Status.DecisionTime.String()), cond.Message)
			},
		},
		"approvalRejected": {
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run, handler *ApprovalRunHandler) {
				tt, err := time.Parse("2006-01-02 15:04:05 -0700 MST", "2021-07-23 14:40:30 +0900 KST")
				require.NoError(t, err)

				require.NoError(t, createApprovalForRun(handler.Client, cicdv1.ApprovalStatus{
					Result:       cicdv1.ApprovalResultRejected,
					Approver:     "admin",
					Reason:       "no!",
					DecisionTime: &metav1.Time{Time: tt},
				}))
			},
			verifyFunc: func(t *testing.T, run *tektonv1alpha1.Run, approval *cicdv1.Approval) {
				cond := run.Status.GetCondition(apis.ConditionSucceeded)
				require.NotNil(t, cond)
				require.True(t, cond.IsFalse())
				require.Equal(t, "Rejected", cond.Reason)
				require.Equal(t, fmt.Sprintf("admin rejected this approval, reason: no!, decisionTime: %s", approval.Status.DecisionTime.String()), cond.Message)
			},
		},
		"alreadyCompleted": {
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run, handler *ApprovalRunHandler) {
				run.Status.SetCondition(&apis.Condition{
					Type:   apis.ConditionSucceeded,
					Status: corev1.ConditionTrue,
				})
				run.Status.CompletionTime = &metav1.Time{Time: time.Now()}
			},
			verifyFunc: func(t *testing.T, run *tektonv1alpha1.Run, approval *cicdv1.Approval) {
				cond := run.Status.GetCondition(apis.ConditionSucceeded)
				require.NotNil(t, cond)
				require.True(t, cond.IsTrue())
				require.NotNil(t, run.Status.CompletionTime)
				require.Equal(t, "", approval.Name)
			},
		},
		"approvalReqMailError": {
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run, handler *ApprovalRunHandler) {
				approvalStatus := cicdv1.ApprovalStatus{Result: cicdv1.ApprovalResultAwaiting}
				meta.SetStatusCondition(&approvalStatus.Conditions, metav1.Condition{
					Type:    cicdv1.ApprovalConditionSentRequestMail,
					Status:  metav1.ConditionFalse,
					Reason:  "ErrorSendingMail",
					Message: "some smtp-related error message!",
				})
				require.NoError(t, createApprovalForRun(handler.Client, approvalStatus))
			},
			verifyFunc: func(t *testing.T, run *tektonv1alpha1.Run, approval *cicdv1.Approval) {
				cond := run.Status.GetCondition(apis.ConditionSucceeded)
				require.NotNil(t, cond)
				require.True(t, cond.IsUnknown())
				require.Equal(t, "Awaiting", cond.Reason)
				require.Equal(t, "RequestMail : ErrorSendingMail-some smtp-related error message!", cond.Message)
			},
		},
		"approvalResMailError": {
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run, handler *ApprovalRunHandler) {
				approvalStatus := cicdv1.ApprovalStatus{Result: cicdv1.ApprovalResultAwaiting}
				meta.SetStatusCondition(&approvalStatus.Conditions, metav1.Condition{
					Type:    cicdv1.ApprovalConditionSentResultMail,
					Status:  metav1.ConditionFalse,
					Reason:  "ErrorSendingMail",
					Message: "some smtp-related error message!",
				})
				require.NoError(t, createApprovalForRun(handler.Client, approvalStatus))
			},
			verifyFunc: func(t *testing.T, run *tektonv1alpha1.Run, approval *cicdv1.Approval) {
				cond := run.Status.GetCondition(apis.ConditionSucceeded)
				require.NotNil(t, cond)
				require.True(t, cond.IsUnknown())
				require.Equal(t, "Awaiting", cond.Reason)
				require.Equal(t, "ResultMail : ErrorSendingMail-some smtp-related error message!", cond.Message)
			},
		},
		"malformedRun": {
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run, handler *ApprovalRunHandler) {
				run.Spec.Params = nil
			},
			verifyFunc: func(t *testing.T, run *tektonv1alpha1.Run, approval *cicdv1.Approval) {
				cond := run.Status.GetCondition(apis.ConditionSucceeded)
				require.NotNil(t, cond)
				require.True(t, cond.IsFalse())
				require.Equal(t, "CannotCreateApproval", cond.Reason)
				require.Equal(t, "there is no param approvers for Run", cond.Message)
			},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			configs.EnableMail = true

			s := runtime.NewScheme()
			utilruntime.Must(corev1.AddToScheme(s))
			utilruntime.Must(cicdv1.AddToScheme(s))
			utilruntime.Must(tektonv1alpha1.AddToScheme(s))

			fakeCli := fake.NewClientBuilder().WithScheme(s).Build()

			handler := &ApprovalRunHandler{
				Client: fakeCli,
				Log:    ctrl.Log.WithName("controllers").WithName("ApprovalRun"),
				Scheme: s,
			}

			// Init Run
			run := generateApprovalRun()
			require.NoError(t, fakeCli.Create(context.Background(), run))
			require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: testApprovalRunName, Namespace: testApprovalRunNamespace}, run))
			if c.preFunc != nil {
				c.preFunc(t, run, handler)
			}
			require.NoError(t, fakeCli.Update(context.Background(), run))
			require.NoError(t, fakeCli.Status().Update(context.Background(), run))

			_, err := handler.Handle(run)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: testApprovalRunName, Namespace: testApprovalRunNamespace}, run))

				approval := &cicdv1.Approval{}
				_ = fakeCli.Get(context.Background(), types.NamespacedName{Name: testApprovalRunName, Namespace: testApprovalRunNamespace}, approval)
				c.verifyFunc(t, run, approval)
			}

			// Delete run, approval
			_ = fakeCli.Delete(context.Background(), &tektonv1alpha1.Run{ObjectMeta: metav1.ObjectMeta{Name: testApprovalRunName, Namespace: testApprovalRunNamespace}})
			_ = fakeCli.Delete(context.Background(), &cicdv1.Approval{ObjectMeta: metav1.ObjectMeta{Name: testApprovalRunName, Namespace: testApprovalRunNamespace}})
		})
	}
}

func TestApprovalRunHandler_reflectApprovalEmailStatus(t *testing.T) {
	tc := map[string]struct {
		approvalConditions []metav1.Condition
		expectedMessage    string
	}{
		"bothFail": {
			approvalConditions: []metav1.Condition{
				{Type: cicdv1.ApprovalConditionSentRequestMail, Status: metav1.ConditionFalse, Reason: "smtp-error", Message: "smtp 403 error!"},
				{Type: cicdv1.ApprovalConditionSentResultMail, Status: metav1.ConditionFalse, Reason: "receiver-error", Message: "receiver is not valid"},
			},
			expectedMessage: "RequestMail : smtp-error-smtp 403 error!, ResultMail : receiver-error-receiver is not valid",
		},
		"reqFail": {
			approvalConditions: []metav1.Condition{
				{Type: cicdv1.ApprovalConditionSentRequestMail, Status: metav1.ConditionFalse, Reason: "smtp-error", Message: "smtp 403 error!"},
				{Type: cicdv1.ApprovalConditionSentResultMail, Status: metav1.ConditionTrue},
			},
			expectedMessage: "RequestMail : smtp-error-smtp 403 error!",
		},
		"resFail": {
			approvalConditions: []metav1.Condition{
				{Type: cicdv1.ApprovalConditionSentRequestMail, Status: metav1.ConditionTrue},
				{Type: cicdv1.ApprovalConditionSentResultMail, Status: metav1.ConditionFalse, Reason: "receiver-error", Message: "receiver is not valid"},
			},
			expectedMessage: "ResultMail : receiver-error-receiver is not valid",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			handler := &ApprovalRunHandler{}
			cond := &apis.Condition{}
			handler.reflectApprovalEmailStatus(&cicdv1.Approval{Status: cicdv1.ApprovalStatus{Conditions: c.approvalConditions}}, cond)
			require.Equal(t, c.expectedMessage, cond.Message)
		})
	}
}

func TestApprovalRunHandler_setApprovalRunStatus(t *testing.T) {
	tc := map[string]struct {
		cond    *apis.Condition
		status  corev1.ConditionStatus
		reason  string
		message string
	}{
		"nil": {
			cond: nil,
		},
		"notNil": {
			cond:    &apis.Condition{},
			status:  corev1.ConditionTrue,
			reason:  "reason1",
			message: "message1",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			handler := &ApprovalRunHandler{}
			handler.setApprovalRunStatus(c.cond, c.status, c.reason, c.message)

			if c.cond == nil {
				require.Nil(t, c.cond)
			} else {
				require.NotNil(t, c.cond)
				require.Equal(t, c.status, c.cond.Status)
				require.Equal(t, c.reason, c.cond.Reason)
				require.Equal(t, c.message, c.cond.Message)
			}
		})
	}
}

func TestApprovalRunHandler_newApproval(t *testing.T) {
	tc := map[string]struct {
		params []tektonv1beta1.Param

		expectedApprovalSpec cicdv1.ApprovalSpec
		errorOccurs          bool
		errorMessage         string
	}{
		"normal": {
			params: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskApprovalParamKeyApprovers, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: []string{"admin1=admin@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskApprovalParamKeyApproversCM, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-cm"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyMessage, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-message"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyIntegrationJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-ij"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyIntegrationJobJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-job"}},
				{Name: cicdv1.CustomTaskApprovalParamKeySenderName, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-sender"}},
				{Name: cicdv1.CustomTaskApprovalParamKeySenderEmail, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-email@tmax.co.kr"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyLink, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-link"}},
			},
			expectedApprovalSpec: cicdv1.ApprovalSpec{
				SkipSendMail:   false,
				PipelineRun:    "pr-1",
				IntegrationJob: "test-ij",
				JobName:        "test-job",
				Message:        "test-message",
				Sender: &cicdv1.ApprovalUser{
					Name:  "test-sender",
					Email: "test-email@tmax.co.kr",
				},
				Link: "test-link",
				Users: []cicdv1.ApprovalUser{
					{Name: "admin1", Email: "admin@tmax.co.kr"},
					{Name: "admin"},
					{Name: "admin2", Email: "admin2@tmax.co.kr"},
				},
			},
		},
		"noLink": {
			params: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskApprovalParamKeyApprovers, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: []string{"admin1=admin@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskApprovalParamKeyApproversCM, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-cm"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyMessage, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-message"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyIntegrationJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-ij"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyIntegrationJobJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-job"}},
				{Name: cicdv1.CustomTaskApprovalParamKeySenderName, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-sender"}},
				{Name: cicdv1.CustomTaskApprovalParamKeySenderEmail, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-email@tmax.co.kr"}},
			},
			errorOccurs:  true,
			errorMessage: "there is no param link for Run",
		},
		"noSenderName": {
			params: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskApprovalParamKeyApprovers, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: []string{"admin1=admin@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskApprovalParamKeyApproversCM, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-cm"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyMessage, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-message"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyIntegrationJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-ij"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyIntegrationJobJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-job"}},
			},
			errorOccurs:  true,
			errorMessage: "there is no param senderName for Run",
		},
		"noJobName": {
			params: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskApprovalParamKeyApprovers, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: []string{"admin1=admin@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskApprovalParamKeyApproversCM, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-cm"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyMessage, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-message"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyIntegrationJob, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-ij"}},
			},
			errorOccurs:  true,
			errorMessage: "there is no param integration-job-job for Run",
		},
		"noIJName": {
			params: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskApprovalParamKeyApprovers, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: []string{"admin1=admin@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskApprovalParamKeyApproversCM, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-cm"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyMessage, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-message"}},
			},
			errorOccurs:  true,
			errorMessage: "there is no param integration-job for Run",
		},
		"noMessage": {
			params: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskApprovalParamKeyApprovers, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: []string{"admin1=admin@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskApprovalParamKeyApproversCM, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-cm"}},
			},
			errorOccurs:  true,
			errorMessage: "there is no param message for Run",
		},
		"noCM": {
			params: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskApprovalParamKeyApprovers, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: []string{"admin1=admin@tmax.co.kr"}}},
			},
			errorOccurs:  true,
			errorMessage: "there is no param approvers-config-map for Run",
		},
		"noApprovers": {
			errorOccurs:  true,
			errorMessage: "there is no param approvers for Run",
		},
		"wrongCM": {
			params: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskApprovalParamKeyApprovers, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: []string{"admin1=admin@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskApprovalParamKeyApproversCM, Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "test-cm-no"}},
			},
			errorOccurs:  true,
			errorMessage: "there is no param message for Run",
		},
	}

	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))
	utilruntime.Must(tektonv1alpha1.AddToScheme(s))

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cm",
			Namespace: "test-ns",
		},
		Data: map[string]string{
			cicdv1.CustomTaskApprovalApproversConfigMapKey: "admin\nadmin2=admin2@tmax.co.kr",
		},
	}

	configMapFail := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cm-fail",
			Namespace: "test-ns",
		},
		Data: map[string]string{
			cicdv1.CustomTaskApprovalApproversConfigMapKey: "admin\nadmin2==@@",
		},
	}

	configMapNoKey := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cm-no",
			Namespace: "test-ns",
		},
		Data: map[string]string{
			"": "admin\nadmin2==@@",
		},
	}

	fakeCli := fake.NewClientBuilder().WithScheme(s).WithObjects(configMap, configMapFail, configMapNoKey).Build()
	handler := &ApprovalRunHandler{Client: fakeCli}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			run := &tektonv1alpha1.Run{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "run-1",
					Namespace: "test-ns",
					OwnerReferences: []metav1.OwnerReference{
						{APIVersion: "", Kind: "PipelineRun", Name: "pr-1"},
						{APIVersion: "tekton.dev/v1beta1", Kind: "PipelineRun", Name: "pr-1"},
					},
				},
				Spec: tektonv1alpha1.RunSpec{
					Params: c.params,
				},
			}
			configs.EnableMail = true
			approval, err := handler.newApproval(run)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedApprovalSpec, approval.Spec)
			}
		})
	}
}

func Test_searchParam(t *testing.T) {
	tc := map[string]struct {
		params       []tektonv1beta1.Param
		key          string
		expectedKind tektonv1beta1.ParamType

		expectedString string
		expectedArray  []string
		errorOccurs    bool
		errorMessage   string
	}{
		"string": {
			params: []tektonv1beta1.Param{
				{Name: "key1", Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "val1"}},
				{Name: "key2", Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: []string{"val2", "val3"}}},
			},
			key:            "key1",
			expectedKind:   tektonv1beta1.ParamTypeString,
			expectedString: "val1",
		},
		"array": {
			params: []tektonv1beta1.Param{
				{Name: "key1", Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "val1"}},
				{Name: "key2", Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: []string{"val2", "val3"}}},
			},
			key:           "key2",
			expectedKind:  tektonv1beta1.ParamTypeArray,
			expectedArray: []string{"val2", "val3"},
		},
		"wrongType": {
			params: []tektonv1beta1.Param{
				{Name: "key1", Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "val1"}},
				{Name: "key2", Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: []string{"val2", "val3"}}},
			},
			key:          "key2",
			expectedKind: tektonv1beta1.ParamTypeString,
			errorOccurs:  true,
			errorMessage: "parameter key2 is expected to be string ,got array",
		},
		"noParam": {
			params: []tektonv1beta1.Param{
				{Name: "key1", Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeString, StringVal: "val1"}},
				{Name: "key2", Value: tektonv1beta1.ArrayOrString{Type: tektonv1beta1.ParamTypeArray, ArrayVal: []string{"val2", "val3"}}},
			},
			key:          "key3",
			errorOccurs:  true,
			errorMessage: "there is no param key3 for Run",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			str, arr, err := searchParam(c.params, c.key, c.expectedKind)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				if c.expectedKind == tektonv1beta1.ParamTypeString {
					require.Equal(t, c.expectedString, str)
				} else {
					require.Equal(t, c.expectedArray, arr)
				}
			}
		})
	}
}

func generateApprovalRun() *tektonv1alpha1.Run {
	return &tektonv1alpha1.Run{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testApprovalRunName,
			Namespace: testApprovalRunNamespace,
		},
		Spec: tektonv1alpha1.RunSpec{
			Params: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskApprovalParamKeyApprovers, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeArray, ArrayVal: []string{"admin@tmax.co.kr=admin@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskApprovalParamKeyApproversCM, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: ""}},
				{Name: cicdv1.CustomTaskApprovalParamKeyMessage, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "approval message!"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyIntegrationJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "ij-sample"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyIntegrationJobJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "job-test"}},
				{Name: cicdv1.CustomTaskApprovalParamKeySenderName, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "developer1"}},
				{Name: cicdv1.CustomTaskApprovalParamKeySenderEmail, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "dev@tmax.co.kr"}},
				{Name: cicdv1.CustomTaskApprovalParamKeyLink, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "https://approval.ref"}},
			},
		},
	}
}

func createApprovalForRun(cli client.Client, status cicdv1.ApprovalStatus) error {
	approval := &cicdv1.Approval{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testApprovalRunName,
			Namespace: testApprovalRunNamespace,
		},
		Spec: cicdv1.ApprovalSpec{},
	}
	if err := cli.Create(context.Background(), approval); err != nil {
		return err
	}
	approval.Status = status
	return cli.Status().Update(context.Background(), approval)
}
