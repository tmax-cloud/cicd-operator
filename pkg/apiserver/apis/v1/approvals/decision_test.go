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

package approvals

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_handler_approveHandler(t *testing.T) {
	s := runtime.NewScheme()
	require.NoError(t, cicdv1.AddToScheme(s))
	approval := &cicdv1.Approval{
		ObjectMeta: metav1.ObjectMeta{Name: "test-approval", Namespace: "test-ns"},
		Spec: cicdv1.ApprovalSpec{
			Users: []cicdv1.ApprovalUser{
				{Name: "test-user"},
			},
		},
	}

	fakeCli := fake.NewClientBuilder().WithScheme(s).WithObjects(approval).Build()
	h := &handler{log: &test.FakeLogger{}, k8sClient: fakeCli}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer([]byte(`{"reason": "test-reason"}`)))
	req = mux.SetURLVars(req, map[string]string{
		"namespace":    "test-ns",
		"approvalName": "test-approval",
	})
	req.Header = map[string][]string{
		"X-Remote-User":  {"test-user"},
		"X-Remote-Group": {"test-group"},
	}
	h.approveHandler(w, req)

	require.Equal(t, 200, w.Result().StatusCode)
	b, err := ioutil.ReadAll(w.Result().Body)
	require.NoError(t, err)
	require.Equal(t, string(b), "{}")

	result := &cicdv1.Approval{}
	require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: "test-approval", Namespace: "test-ns"}, result))
	require.Equal(t, cicdv1.ApprovalResultApproved, result.Status.Result)
	require.Equal(t, "test-reason", result.Status.Reason)
	require.Equal(t, "test-user", result.Status.Approver)
}

func Test_handler_rejectHandler(t *testing.T) {
	s := runtime.NewScheme()
	require.NoError(t, cicdv1.AddToScheme(s))
	approval := &cicdv1.Approval{
		ObjectMeta: metav1.ObjectMeta{Name: "test-approval", Namespace: "test-ns"},
		Spec: cicdv1.ApprovalSpec{
			Users: []cicdv1.ApprovalUser{
				{Name: "test-user"},
			},
		},
	}

	fakeCli := fake.NewClientBuilder().WithScheme(s).WithObjects(approval).Build()
	h := &handler{log: &test.FakeLogger{}, k8sClient: fakeCli}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer([]byte(`{"reason": "test-reason"}`)))
	req = mux.SetURLVars(req, map[string]string{
		"namespace":    "test-ns",
		"approvalName": "test-approval",
	})
	req.Header = map[string][]string{
		"X-Remote-User":  {"test-user"},
		"X-Remote-Group": {"test-group"},
	}
	h.rejectHandler(w, req)

	require.Equal(t, 200, w.Result().StatusCode)
	b, err := ioutil.ReadAll(w.Result().Body)
	require.NoError(t, err)
	require.Equal(t, string(b), "{}")

	result := &cicdv1.Approval{}
	require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: "test-approval", Namespace: "test-ns"}, result))
	require.Equal(t, cicdv1.ApprovalResultRejected, result.Status.Result)
	require.Equal(t, "test-reason", result.Status.Reason)
	require.Equal(t, "test-user", result.Status.Approver)
}

func Test_handler_updateDecision(t *testing.T) {
	s := runtime.NewScheme()
	require.NoError(t, cicdv1.AddToScheme(s))

	tc := map[string]struct {
		decision cicdv1.ApprovalResult
		body     io.Reader
		vars     map[string]string
		header   http.Header
		approval *cicdv1.Approval

		expectedCode    int
		expectedMessage string
		expectedResult  cicdv1.ApprovalResult
		expectedReason  string
		expectedUser    string
	}{
		"normal": {
			decision: cicdv1.ApprovalResultApproved,
			body:     bytes.NewBuffer([]byte(`{"reason": "test-reason"}`)),
			vars: map[string]string{
				"namespace":    "test-ns",
				"approvalName": "test-approval",
			},
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{Name: "test-approval", Namespace: "test-ns"},
				Spec: cicdv1.ApprovalSpec{
					Users: []cicdv1.ApprovalUser{
						{Name: "test-user"},
					},
				},
			},
			expectedCode:    200,
			expectedMessage: "{}",
			expectedResult:  cicdv1.ApprovalResultApproved,
			expectedReason:  "test-reason",
			expectedUser:    "test-user",
		},
		"noParam": {
			decision: cicdv1.ApprovalResultApproved,
			body:     bytes.NewBuffer([]byte(`{"reason": "test-reason"}`)),
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{Name: "test-approval", Namespace: "test-ns"},
				Spec: cicdv1.ApprovalSpec{
					Users: []cicdv1.ApprovalUser{
						{Name: "test-user"},
					},
				},
			},
			expectedCode:    400,
			expectedMessage: "{\"message\":\"url is malformed\"}",
		},
		"decodeBodyErr": {
			decision: cicdv1.ApprovalResultApproved,
			body:     bytes.NewBuffer([]byte(`{"reason": "test-reason`)),
			vars: map[string]string{
				"namespace":    "test-ns",
				"approvalName": "test-approval",
			},
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{Name: "test-approval", Namespace: "test-ns"},
				Spec: cicdv1.ApprovalSpec{
					Users: []cicdv1.ApprovalUser{
						{Name: "test-user"},
					},
				},
			},
			expectedCode:    400,
			expectedMessage: "body is not in json form or is malformed",
		},
		"noUserHeader": {
			decision: cicdv1.ApprovalResultApproved,
			body:     bytes.NewBuffer([]byte(`{"reason": "test-reason"}`)),
			vars: map[string]string{
				"namespace":    "test-ns",
				"approvalName": "test-approval",
			},
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{Name: "test-approval", Namespace: "test-ns"},
				Spec: cicdv1.ApprovalSpec{
					Users: []cicdv1.ApprovalUser{
						{Name: "test-user"},
					},
				},
			},
			expectedCode:    401,
			expectedMessage: "forbidden user, err : no header X-Remote-User",
		},
		"getApprovalErr": {
			decision: cicdv1.ApprovalResultApproved,
			body:     bytes.NewBuffer([]byte(`{"reason": "test-reason"}`)),
			vars: map[string]string{
				"namespace":    "test-ns",
				"approvalName": "test-approval",
			},
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			expectedCode:    400,
			expectedMessage: "no Approval test-ns/test-approval is found",
		},
		"alreadyDecided": {
			decision: cicdv1.ApprovalResultApproved,
			body:     bytes.NewBuffer([]byte(`{"reason": "test-reason"}`)),
			vars: map[string]string{
				"namespace":    "test-ns",
				"approvalName": "test-approval",
			},
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{Name: "test-approval", Namespace: "test-ns"},
				Spec: cicdv1.ApprovalSpec{
					Users: []cicdv1.ApprovalUser{
						{Name: "test-user"},
					},
				},
				Status: cicdv1.ApprovalStatus{
					Result: cicdv1.ApprovalResultRejected,
				},
			},
			expectedCode:    400,
			expectedMessage: "approval test-ns/test-approval is already in Rejected status",
		},
		"notValidApprover": {
			decision: cicdv1.ApprovalResultApproved,
			body:     bytes.NewBuffer([]byte(`{"reason": "test-reason"}`)),
			vars: map[string]string{
				"namespace":    "test-ns",
				"approvalName": "test-approval",
			},
			header: map[string][]string{
				"X-Remote-User":  {"test-user2"},
				"X-Remote-Group": {"test-group"},
			},
			approval: &cicdv1.Approval{
				ObjectMeta: metav1.ObjectMeta{Name: "test-approval", Namespace: "test-ns"},
				Spec: cicdv1.ApprovalSpec{
					Users: []cicdv1.ApprovalUser{
						{Name: "test-user"},
					},
				},
			},
			expectedCode:    403,
			expectedMessage: "approval test-ns/test-approval is not requested to you",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			fakeCli := fake.NewClientBuilder().WithScheme(s).Build()
			if c.approval != nil {
				require.NoError(t, fakeCli.Create(context.Background(), c.approval))
			}

			h := &handler{log: &test.FakeLogger{}, k8sClient: fakeCli}

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/", c.body)
			req = mux.SetURLVars(req, c.vars)
			req.Header = c.header
			h.updateDecision(w, req, c.decision)

			require.Equal(t, c.expectedCode, w.Result().StatusCode)
			b, err := ioutil.ReadAll(w.Result().Body)
			require.NoError(t, err)
			require.Contains(t, string(b), c.expectedMessage)

			if c.expectedCode == 200 {
				result := &cicdv1.Approval{}
				require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: c.approval.Name, Namespace: c.approval.Namespace}, result))
				require.Equal(t, c.expectedResult, result.Status.Result)
				require.Equal(t, c.expectedReason, result.Status.Reason)
				require.Equal(t, c.expectedUser, result.Status.Approver)
			}
		})
	}
}
