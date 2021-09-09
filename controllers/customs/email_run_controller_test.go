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
	"net"
	"net/textproto"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
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
	testEmailRunName = "test-email-run"
)

type testEmailStruct struct {
	from   string
	to     []string
	header map[string]string
	data   []string
}

var testEmailResult testEmailStruct

func TestEmailRunHandler_Handle(t *testing.T) {
	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(cicdv1.AddToScheme(s))
	utilruntime.Must(tektonv1beta1.AddToScheme(s))
	utilruntime.Must(tektonv1alpha1.AddToScheme(s))

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() {
		_ = l.Close()
	}()

	configs.SMTPHost = l.Addr().String()

	tc := map[string]struct {
		runParams    []tektonv1beta1.Param
		runCondition *apis.Condition
		disableEmail bool

		expectedEmail testEmailStruct
		expectedCond  *apis.Condition
	}{
		"success": {
			runParams: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskEmailParamKeyReceivers, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeArray, ArrayVal: []string{"re@tmax.co.kr", "re2@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskEmailParamKeyTitle, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME/$JOB_NAME"}},
				{Name: cicdv1.CustomTaskEmailParamKeyContent, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME - $JOB_NAME"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIsHTML, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "false"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-ij-1"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJobJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-job-1"}},
			},
			expectedCond: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionTrue,
				Reason:  "SentMail",
				Message: "",
			},
			expectedEmail: testEmailStruct{
				from: "FROM:<admin@tmax.co.kr>",
				to:   []string{"TO:<re@tmax.co.kr>", "TO:<re2@tmax.co.kr>"},
				header: map[string]string{
					"From":         "admin@tmax.co.kr",
					"To":           "<re@tmax.co.kr>, <re2@tmax.co.kr>",
					"Subject":      "test-ij-1/test-job-1",
					"MIME-Version": "1.0",
					"Content-Type": "text/plain; charset=UTF-8",
				},
				data: []string{"test-ij-1 - test-job-1"},
			},
		},
		"alreadyCompleted": {
			runParams: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskEmailParamKeyReceivers, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeArray, ArrayVal: []string{"re@tmax.co.kr", "re2@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskEmailParamKeyTitle, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME/$JOB_NAME"}},
				{Name: cicdv1.CustomTaskEmailParamKeyContent, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME - $JOB_NAME"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIsHTML, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "false"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-ij-1"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJobJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-job-1"}},
			},
			runCondition: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionTrue,
				Reason:  "SentMail",
				Message: "",
			},
			expectedCond: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionTrue,
				Reason:  "SentMail",
				Message: "",
			},
		},
		"noReceivers": {
			runParams: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskEmailParamKeyTitle, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME/$JOB_NAME"}},
				{Name: cicdv1.CustomTaskEmailParamKeyContent, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME - $JOB_NAME"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIsHTML, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "false"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-ij-1"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJobJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-job-1"}},
			},
			expectedCond: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionFalse,
				Reason:  "InsufficientParams",
				Message: "there is no param receivers for Run",
			},
		},
		"noContent": {
			runParams: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskEmailParamKeyReceivers, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeArray, ArrayVal: []string{"re@tmax.co.kr", "re2@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskEmailParamKeyTitle, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME/$JOB_NAME"}},
			},
			expectedCond: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionFalse,
				Reason:  "InsufficientParams",
				Message: "there is no param content for Run",
			},
		},
		"noTitle": {
			runParams: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskEmailParamKeyReceivers, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeArray, ArrayVal: []string{"re@tmax.co.kr", "re2@tmax.co.kr"}}},
			},
			expectedCond: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionFalse,
				Reason:  "InsufficientParams",
				Message: "there is no param title for Run",
			},
		},
		"isHTML": {
			runParams: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskEmailParamKeyReceivers, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeArray, ArrayVal: []string{"re@tmax.co.kr", "re2@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskEmailParamKeyTitle, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME/$JOB_NAME"}},
				{Name: cicdv1.CustomTaskEmailParamKeyContent, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME - $JOB_NAME"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIsHTML, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "true"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-ij-1"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJobJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-job-1"}},
			},
			expectedCond: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionTrue,
				Reason:  "SentMail",
				Message: "",
			},
			expectedEmail: testEmailStruct{
				from: "FROM:<admin@tmax.co.kr>",
				to:   []string{"TO:<re@tmax.co.kr>", "TO:<re2@tmax.co.kr>"},
				header: map[string]string{
					"From":         "admin@tmax.co.kr",
					"To":           "<re@tmax.co.kr>, <re2@tmax.co.kr>",
					"Subject":      "test-ij-1/test-job-1",
					"MIME-Version": "1.0",
					"Content-Type": "text/html; charset=UTF-8",
				},
				data: []string{"test-ij-1 - test-job-1"},
			},
		},
		"compileTitleError": {
			runParams: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskEmailParamKeyReceivers, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeArray, ArrayVal: []string{"re@tmax.co.kr", "re2@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskEmailParamKeyTitle, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "{{...}}"}},
				{Name: cicdv1.CustomTaskEmailParamKeyContent, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME - $JOB_NAME"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIsHTML, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "true"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-ij-1"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJobJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-job-1"}},
			},
			expectedCond: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionFalse,
				Reason:  "CannotCompileTitle",
				Message: "template: contentTemplate:1: unexpected <.> in operand",
			},
		},
		"compileMessageError": {
			runParams: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskEmailParamKeyReceivers, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeArray, ArrayVal: []string{"re@tmax.co.kr", "re2@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskEmailParamKeyTitle, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME/$JOB_NAME"}},
				{Name: cicdv1.CustomTaskEmailParamKeyContent, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "{{...}}"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIsHTML, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "false"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-ij-1"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJobJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-job-1"}},
			},
			expectedCond: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionFalse,
				Reason:  "CannotCompileContent",
				Message: "template: contentTemplate:1: unexpected <.> in operand",
			},
		},
		"sendError": {
			disableEmail: true,
			runParams: []tektonv1beta1.Param{
				{Name: cicdv1.CustomTaskEmailParamKeyReceivers, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeArray, ArrayVal: []string{"re@tmax.co.kr", "re2@tmax.co.kr"}}},
				{Name: cicdv1.CustomTaskEmailParamKeyTitle, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME/$JOB_NAME"}},
				{Name: cicdv1.CustomTaskEmailParamKeyContent, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "$INTEGRATION_JOB_NAME - $JOB_NAME"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIsHTML, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "false"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-ij-1"}},
				{Name: cicdv1.CustomTaskEmailParamKeyIntegrationJobJob, Value: tektonv1alpha1.ArrayOrString{Type: tektonv1alpha1.ParamTypeString, StringVal: "test-job-1"}},
			},
			expectedCond: &apis.Condition{
				Type:    apis.ConditionSucceeded,
				Status:  corev1.ConditionFalse,
				Reason:  "EmailError",
				Message: "email is disabled",
			},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			testEmailResult = testEmailStruct{}
			configs.EnableMail = !c.disableEmail
			configs.SMTPUserSecret = "smtp-auth"

			exitCh := make(chan struct{}, 1)

			go mockSMTPServer(l, t, exitCh)

			ns, err := utils.Namespace()
			require.NoError(t, err)

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "smtp-auth",
					Namespace: ns,
				},
				Type: corev1.SecretTypeBasicAuth,
				Data: map[string][]byte{
					corev1.BasicAuthUsernameKey: []byte("admin@tmax.co.kr"),
					corev1.BasicAuthPasswordKey: []byte("admin"),
				},
			}

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
					Name:      testEmailRunName,
					Namespace: testRunNamespace,
				},
				Spec: tektonv1alpha1.RunSpec{
					Ref: &tektonv1alpha1.TaskRef{
						APIVersion: "cicd.tmax.io/v1",
						Kind:       "EmailTask",
					},
					Params: c.runParams,
				},
				Status: tektonv1alpha1.RunStatus{},
			}
			if c.runCondition != nil {
				run.Status.Conditions = append(run.Status.Conditions, *c.runCondition)
			}

			fakeCli := fake.NewFakeClientWithScheme(s, run, ij, secret)
			emailRunHandler := EmailRunHandler{Scheme: s, Client: fakeCli, Log: ctrl.Log}

			res, err := emailRunHandler.Handle(run)
			require.NoError(t, err)
			require.Equal(t, ctrl.Result{}, res)

			resRun := &tektonv1alpha1.Run{}
			require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: testEmailRunName, Namespace: testRunNamespace}, resRun))

			cond := resRun.Status.GetCondition(apis.ConditionSucceeded)
			if c.expectedCond != nil {
				require.NotNil(t, cond)
				require.Equal(t, c.expectedCond.Status, cond.Status)
				require.Equal(t, c.expectedCond.Reason, cond.Reason)
				require.Equal(t, c.expectedCond.Message, cond.Message)
				if c.expectedCond.Status == corev1.ConditionTrue {
					require.Equal(t, c.expectedEmail, testEmailResult)
				}
			}
			exitCh <- struct{}{}
		})
	}
}

func mockSMTPServer(l net.Listener, t *testing.T, exitCh chan struct{}) {
	conn, err := l.Accept()
	if err != nil {
		for {
			select {
			case <-exitCh:
				return
			default:
				t.Errorf("accept: %v", err)
				return
			}
		}
	}
	defer func() {
		_ = conn.Close()
	}()

	tc := textproto.NewConn(conn)
	require.NoError(t, tc.PrintfLine("220 hello world"))

	msg, err := tc.ReadLine()
	require.NoError(t, err)
	require.Equal(t, "EHLO localhost", msg)

	require.NoError(t, tc.PrintfLine("Hello localhost"))
	require.NoError(t, tc.PrintfLine("250 AUTH LOGIN PLAIN"))

	msg, err = tc.ReadLine()
	require.NoError(t, err)
	require.Equal(t, "HELO localhost", msg)

	require.NoError(t, tc.PrintfLine("250 AUTH LOGIN PLAIN"))

	isData := false
	isHeader := false
	for {
		id := tc.Next()

		msg, err = tc.ReadLine()
		require.NoError(t, err)
		t.Logf("REQ: %s\n", msg)

		if isData {
			handleData(t, tc, id, msg, &isData, &isHeader)
			continue
		}

		tc.StartResponse(id)
		cmd := msg[:4]
		var resp string
		switch cmd {
		case "MAIL":
			testEmailResult.from = msg[5:]
		case "RCPT":
			testEmailResult.to = append(testEmailResult.to, msg[5:])
			resp = "250 OK"
		case "DATA":
			isData = true
			isHeader = true
			testEmailResult.header = map[string]string{}
			resp = "354 Go ahead"
		case "QUIT":
			resp = "221 Good bye"
		}

		respond(t, tc, resp)
		tc.EndResponse(id)
		if cmd == "QUIT" {
			break
		}
	}
}

func handleData(t *testing.T, tc *textproto.Conn, id uint, msg string, isData, isHeader *bool) {
	tc.StartResponse(id)
	switch msg {
	case ".":
		respond(t, tc, "250 OK")
		*isData = false
	case "":
		if *isHeader {
			*isHeader = false
		} else {
			testEmailResult.data = append(testEmailResult.data, msg)
		}
	default:
		if *isHeader {
			tok := strings.Split(msg, ": ")
			testEmailResult.header[tok[0]] = tok[1]
		} else {
			testEmailResult.data = append(testEmailResult.data, msg)
		}
	}
	tc.EndResponse(id)
}

func respond(t *testing.T, tc *textproto.Conn, msg string) {
	if msg != "" {
		t.Logf("---> %s\n", msg)
		_ = tc.PrintfLine(msg)
	}
}
