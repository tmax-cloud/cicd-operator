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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/notification/slack"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"knative.dev/pkg/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testSlackRunName = "test-slack-run"
	testRunNamespace = "default"
)

var testSlackMessage slack.Message

func TestSlackRunHandler_Handle(t *testing.T) {
	// Launch test slack webhook server
	srv := newTestServer()
	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))
	utilruntime.Must(tektonv1beta1.AddToScheme(s))
	utilruntime.Must(tektonv1alpha1.AddToScheme(s))

	tc := map[string]struct {
		runParams    []tektonv1beta1.Param
		runCondition *apis.Condition

		expectedMessage slack.Message
		expectedCond    *apis.Condition
	}{
		"normal": {
			runParams: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskSlackParamKeyWebhook, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: srv.URL}},
				{Name: cicdv1.CustomTaskSlackParamKeyMessage, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME/$JOB_NAME"}},
				{Name: cicdv1.CustomTaskSlackParamKeyIntegrationJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-ij-1"}},
				{Name: cicdv1.CustomTaskSlackParamKeyIntegrationJobJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-job-1"}},
			},
			expectedCond: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionTrue,
				Reason:  "SentSlack",
				Message: "",
			},
			expectedMessage: slack.Message{
				Text: "IntegrationJobNotification",
				Blocks: []slack.MessageBlock{
					{Type: "section", Text: slack.BlockText{Type: "mrkdwn", Text: "test-ij-1/test-job-1"}},
				},
			},
		},
		"alreadyCompleted": {
			runParams: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskSlackParamKeyWebhook, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: srv.URL}},
				{Name: cicdv1.CustomTaskSlackParamKeyMessage, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME/$JOB_NAME"}},
				{Name: cicdv1.CustomTaskSlackParamKeyIntegrationJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-ij-1"}},
				{Name: cicdv1.CustomTaskSlackParamKeyIntegrationJobJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-job-1"}},
			},
			runCondition: &apis.Condition{
				Type:   apis.ConditionSucceeded,
				Status: corev1.ConditionFalse,
			},
			expectedCond: &apis.Condition{
				Type:   apis.ConditionSucceeded,
				Status: corev1.ConditionFalse,
			},
		},
		"noSlackWebhook": {
			runParams: []tektonv1beta1.Param{},
			expectedCond: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionFalse,
				Reason:  "InsufficientParams",
				Message: "there is no param webhook-url for Run",
			},
		},
		"noMessage": {
			runParams: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskSlackParamKeyWebhook, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: srv.URL}},
			},
			expectedCond: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionFalse,
				Reason:  "InsufficientParams",
				Message: "there is no param message for Run",
			},
		},
		"compileError": {
			runParams: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskSlackParamKeyWebhook, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: srv.URL}},
				{Name: cicdv1.CustomTaskSlackParamKeyMessage, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "{{..asdsd}}"}},
				{Name: cicdv1.CustomTaskSlackParamKeyIntegrationJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-ij-1"}},
				{Name: cicdv1.CustomTaskSlackParamKeyIntegrationJobJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-job-1"}},
			},
			expectedCond: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionFalse,
				Reason:  "CannotCompileMessage",
				Message: "template: contentTemplate:1: unexpected . after term \".\"",
			},
		},
		"sendError": {
			runParams: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskSlackParamKeyWebhook, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: srv.URL + "/error"}},
				{Name: cicdv1.CustomTaskSlackParamKeyMessage, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME/$JOB_NAME"}},
				{Name: cicdv1.CustomTaskSlackParamKeyIntegrationJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-ij-1"}},
				{Name: cicdv1.CustomTaskSlackParamKeyIntegrationJobJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-job-1"}},
			},
			expectedCond: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionFalse,
				Reason:  "SlackError",
				Message: "status: 400, error: ",
			},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			ij := &cicdv1.IntegrationJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ij-1",
					Namespace: testRunNamespace,
				},
				Spec: cicdv1.IntegrationJobSpec{
					Refs: cicdv1.IntegrationJobRefs{
						Base: cicdv1.IntegrationJobRefsBase{
							Ref: cicdv1.GitRef("refs/tags/v0.2.3"),
						},
					},
					Jobs: []cicdv1.Job{{Container: corev1.Container{Name: "test-job-1"}}},
				},
			}

			run := &tektonv1alpha1.Run{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testSlackRunName,
					Namespace: testRunNamespace,
				},
				Spec: tektonv1alpha1.RunSpec{
					Ref: &tektonv1alpha1.TaskRef{
						APIVersion: "cicd.tmax.io/v1",
						Kind:       "SlackTask",
					},
					Params: c.runParams,
				},
				Status: tektonv1alpha1.RunStatus{},
			}
			if c.runCondition != nil {
				run.Status.Conditions = append(run.Status.Conditions, *c.runCondition)
			}

			fakeCli := fake.NewClientBuilder().WithScheme(s).WithObjects(run, ij).Build()
			slackRunHandler := SlackRunHandler{Scheme: s, Client: fakeCli, Log: ctrl.Log}
			res, err := slackRunHandler.Handle(run)
			require.NoError(t, err)
			require.Equal(t, ctrl.Result{}, res)

			resRun := &tektonv1alpha1.Run{}
			require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: testSlackRunName, Namespace: testRunNamespace}, resRun))

			cond := resRun.Status.GetCondition(apis.ConditionSucceeded)
			if c.expectedCond != nil {
				require.NotNil(t, cond)
				require.Equal(t, c.expectedCond.Status, cond.Status)
				require.Equal(t, c.expectedCond.Reason, cond.Reason)
				require.Equal(t, c.expectedCond.Message, cond.Message)
				if c.expectedCond.Status == corev1.ConditionTrue {
					require.Equal(t, c.expectedMessage, testSlackMessage)
				}
			}
		})
	}
}

func newTestServer() *httptest.Server {
	router := mux.NewRouter()

	router.HandleFunc("/error", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			_ = req.Body.Close()
		}()
		userReq := slack.Message{}
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&userReq); err != nil {
			_ = utils.RespondError(w, http.StatusBadRequest, fmt.Sprintf("body is not in json form or is malformed, err : %s", err.Error()))
			return
		}
		testSlackMessage = userReq
	})

	return httptest.NewServer(router)
}
