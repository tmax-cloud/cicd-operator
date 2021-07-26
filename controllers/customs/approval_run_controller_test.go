package customs

import (
	"context"
	"fmt"
	"github.com/operator-framework/operator-lib/status"
	"github.com/stretchr/testify/require"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"knative.dev/pkg/apis"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
	"time"
)

const (
	testApprovalRunName      = "test-approval"
	testApprovalRunNamespace = "test-ns"
)

type approvalHandlerTestCase struct {
	preFunc    func(t *testing.T, run *tektonv1alpha1.Run)
	verifyFunc func(t *testing.T, run *tektonv1alpha1.Run, approval *cicdv1.Approval)

	errorOccurs  bool
	errorMessage string
}

func TestApprovalRunHandler_Handle(t *testing.T) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))
	utilruntime.Must(tektonv1alpha1.AddToScheme(s))

	fakeCli := fake.NewFakeClientWithScheme(s)

	handler := &ApprovalRunHandler{
		Client: fakeCli,
		Log:    ctrl.Log.WithName("controllers").WithName("ApprovalRun"),
		Scheme: s,
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
				require.Equal(t, "admin@tmax.co.kr=admin@tmax.co.kr", approval.Spec.Users[0])
			},
		},
		"creationNoMail": {
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run) {
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
				require.Equal(t, "admin@tmax.co.kr=admin@tmax.co.kr", approval.Spec.Users[0])
			},
		},
		"creationApproversCM": {
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run) {
				configMapName := "approver-cm-1"
				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: configMapName, Namespace: testApprovalRunNamespace},
					Data:       map[string]string{cicdv1.CustomTaskApprovalApproversConfigMapKey: "admin2@tmax.co.kr=admin2@tmax.co.kr,admin3@tmax.co.kr"},
				}
				require.NoError(t, fakeCli.Create(context.Background(), cm))
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
				require.Equal(t, "admin@tmax.co.kr=admin@tmax.co.kr", approval.Spec.Users[0])
				require.Equal(t, "admin2@tmax.co.kr=admin2@tmax.co.kr", approval.Spec.Users[1])
				require.Equal(t, "admin3@tmax.co.kr", approval.Spec.Users[2])
			},
		},
		"approvalApproved": {
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run) {
				tt, err := time.Parse("2006-01-02 15:04:05 -0700 MST", "2021-07-23 14:40:30 +0900 KST")
				require.NoError(t, err)

				require.NoError(t, createApprovalForRun(fakeCli, cicdv1.ApprovalStatus{
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
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run) {
				tt, err := time.Parse("2006-01-02 15:04:05 -0700 MST", "2021-07-23 14:40:30 +0900 KST")
				require.NoError(t, err)

				require.NoError(t, createApprovalForRun(fakeCli, cicdv1.ApprovalStatus{
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
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run) {
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
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run) {
				approvalStatus := cicdv1.ApprovalStatus{Result: cicdv1.ApprovalResultAwaiting}
				approvalStatus.Conditions.SetCondition(status.Condition{
					Type:    cicdv1.ApprovalConditionSentRequestMail,
					Status:  corev1.ConditionFalse,
					Reason:  "ErrorSendingMail",
					Message: "some smtp-related error message!",
				})
				require.NoError(t, createApprovalForRun(fakeCli, approvalStatus))
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
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run) {
				approvalStatus := cicdv1.ApprovalStatus{Result: cicdv1.ApprovalResultAwaiting}
				approvalStatus.Conditions.SetCondition(status.Condition{
					Type:    cicdv1.ApprovalConditionSentResultMail,
					Status:  corev1.ConditionFalse,
					Reason:  "ErrorSendingMail",
					Message: "some smtp-related error message!",
				})
				require.NoError(t, createApprovalForRun(fakeCli, approvalStatus))
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
			preFunc: func(t *testing.T, run *tektonv1alpha1.Run) {
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

			// Init Run
			run := generateApprovalRun()
			require.NoError(t, fakeCli.Create(context.Background(), run))
			require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: testApprovalRunName, Namespace: testApprovalRunNamespace}, run))
			if c.preFunc != nil {
				c.preFunc(t, run)
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
