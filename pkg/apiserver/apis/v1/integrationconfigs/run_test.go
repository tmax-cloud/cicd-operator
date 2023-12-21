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

package integrationconfigs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/test"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"github.com/tmax-cloud/cicd-operator/pkg/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_handler_runPreHandler(t *testing.T) {
	s := runtime.NewScheme()
	require.NoError(t, cicdv1.AddToScheme(s))

	ic := &cicdv1.IntegrationConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
		Spec: cicdv1.IntegrationConfigSpec{
			Git: cicdv1.GitConfig{
				Type:       "fake",
				APIUrl:     "https://test.git.com",
				Repository: "test/test",
			},
		},
	}

	fakeCli := fake.NewClientBuilder().WithScheme(s).WithObjects(ic).Build()

	h := &handler{log: &test.FakeLogger{}, k8sClient: fakeCli}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer([]byte(`{"base_branch": "master", "head_branch": "feat/test"}`)))
	req = mux.SetURLVars(req, map[string]string{
		"namespace": "test-ns",
		"icName":    "test-ic",
	})
	req.Header = map[string][]string{
		"X-Remote-User":  {"test-user"},
		"X-Remote-Group": {"test-group"},
	}
	h.runPreHandler(w, req)

	require.Equal(t, 200, w.Result().StatusCode)
	b, err := ioutil.ReadAll(w.Result().Body)
	require.NoError(t, err)
	require.Equal(t, "{}", string(b))
}

func Test_handler_runPostHandler(t *testing.T) {
	s := runtime.NewScheme()
	require.NoError(t, cicdv1.AddToScheme(s))

	ic := &cicdv1.IntegrationConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
		Spec: cicdv1.IntegrationConfigSpec{
			Git: cicdv1.GitConfig{
				Type:       "fake",
				APIUrl:     "https://test.git.com",
				Repository: "test/test",
			},
		},
	}

	fakeCli := fake.NewClientBuilder().WithScheme(s).WithObjects(ic).Build()

	h := &handler{log: &test.FakeLogger{}, k8sClient: fakeCli}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer([]byte(`{"base_branch": "master", "head_branch": "feat/test"}`)))
	req = mux.SetURLVars(req, map[string]string{
		"namespace": "test-ns",
		"icName":    "test-ic",
	})
	req.Header = map[string][]string{
		"X-Remote-User":  {"test-user"},
		"X-Remote-Group": {"test-group"},
	}
	h.runPostHandler(w, req)

	require.Equal(t, 200, w.Result().StatusCode)
	b, err := ioutil.ReadAll(w.Result().Body)
	require.NoError(t, err)
	require.Equal(t, "{}", string(b))
}

func Test_handler_runHandler(t *testing.T) {
	s := runtime.NewScheme()
	require.NoError(t, cicdv1.AddToScheme(s))

	server.AddPlugin([]git.EventType{git.EventTypePush, git.EventTypePullRequest}, &testPlugin{})

	tc := map[string]struct {
		event  git.EventType
		body   io.Reader
		vars   map[string]string
		header http.Header
		ic     *cicdv1.IntegrationConfig

		expectedCode    int
		expectedMessage string
	}{
		"pr": {
			event: git.EventTypePullRequest,
			body:  bytes.NewBuffer([]byte(`{"base_branch": "master", "head_branch": "feat/test"}`)),
			vars: map[string]string{
				"namespace": "test-ns",
				"icName":    "test-ic",
			},
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       "fake",
						APIUrl:     "https://test.git.com",
						Repository: "test/test",
					},
				},
			},
			expectedCode:    200,
			expectedMessage: "{}",
		},
		"push": {
			event: git.EventTypePush,
			body:  bytes.NewBuffer([]byte(`{"branch": "master"}`)),
			vars: map[string]string{
				"namespace": "test-ns",
				"icName":    "test-ic",
			},
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       "fake",
						APIUrl:     "https://test.git.com",
						Repository: "test/test",
					},
				},
			},
			expectedCode:    200,
			expectedMessage: "{}",
		},
		"noParam": {
			event: git.EventTypePush,
			body:  bytes.NewBuffer([]byte(`{"branch": "master"}`)),
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       "fake",
						APIUrl:     "https://test.git.com",
						Repository: "test/test",
					},
				},
			},
			expectedCode:    400,
			expectedMessage: "url is malformed",
		},
		"noUserHeader": {
			event: git.EventTypePush,
			body:  bytes.NewBuffer([]byte(`{"branch": "master"}`)),
			vars: map[string]string{
				"namespace": "test-ns",
				"icName":    "test-ic",
			},
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       "fake",
						APIUrl:     "https://test.git.com",
						Repository: "test/test",
					},
				},
			},
			expectedCode:    401,
			expectedMessage: "forbidden user, err : no header X-Remote-User",
		},
		"getICErr": {
			event: git.EventTypePush,
			body:  bytes.NewBuffer([]byte(`{"branch": "master"}`)),
			vars: map[string]string{
				"namespace": "test-ns",
				"icName":    "test-ic",
			},
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			expectedCode:    500,
			expectedMessage: "cannot get IntegrationConfig test-ns/test-ic",
		},
		"getGitHostErr": {
			event: git.EventTypePush,
			body:  bytes.NewBuffer([]byte(`{"branch": "master"}`)),
			vars: map[string]string{
				"namespace": "test-ns",
				"icName":    "test-ic",
			},
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       "fake",
						APIUrl:     "https://192.168.0.%31/",
						Repository: "test/test",
					},
				},
			},
			expectedCode:    500,
			expectedMessage: "cannot get IntegrationConfig test-ns/test-ic's git host",
		},
		"buildPRErr": {
			event: git.EventTypePullRequest,
			body:  bytes.NewBuffer([]byte(`{}`)),
			vars: map[string]string{
				"namespace": "test-ns",
				"icName":    "test-ic",
			},
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       "fake",
						APIUrl:     "https://test.git.com",
						Repository: "test/test",
					},
				},
			},
			expectedCode:    400,
			expectedMessage: "cannot build pull_request webhook",
		},
		"buildPushErr": {
			event: git.EventTypePush,
			body:  bytes.NewBuffer([]byte(`{{{{`)),
			vars: map[string]string{
				"namespace": "test-ns",
				"icName":    "test-ic",
			},
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       "fake",
						APIUrl:     "https://test.git.com",
						Repository: "test/test",
					},
				},
			},
			expectedCode:    400,
			expectedMessage: "cannot build push webhook",
		},
		"triggerErr": {
			event: git.EventTypePush,
			body:  bytes.NewBuffer([]byte(`{"branch": "master"}`)),
			vars: map[string]string{
				"namespace": "test-ns",
				"icName":    "test-err",
			},
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-err", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       "fake",
						APIUrl:     "https://test.git.com",
						Repository: "test/test",
					},
				},
			},
			expectedCode:    500,
			expectedMessage: "cannot handle event, err : test-err returns error",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			fakeCli := fake.NewClientBuilder().WithScheme(s).Build()
			if c.ic != nil {
				require.NoError(t, fakeCli.Create(context.Background(), c.ic))
			}

			h := &handler{log: &test.FakeLogger{}, k8sClient: fakeCli}

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/", c.body)
			req = mux.SetURLVars(req, c.vars)
			req.Header = c.header
			h.runHandler(w, req, c.event)

			require.Equal(t, c.expectedCode, w.Result().StatusCode)
			b, err := ioutil.ReadAll(w.Result().Body)
			require.NoError(t, err)
			require.Contains(t, string(b), c.expectedMessage)
		})
	}
}

type testPlugin struct{}

func (t *testPlugin) Name() string {
	return "dispatcher"
}

func (t *testPlugin) Handle(_ *git.Webhook, config *cicdv1.IntegrationConfig) error {
	fmt.Println(config.Name)
	if config.Name == "test-err" {
		return fmt.Errorf("test-err returns error")
	}
	return nil
}

// func Test_buildPullRequestWebhook(t *testing.T) {
// 	tc := map[string]struct {
// 		body io.Reader

// 		errorOccurs  bool
// 		errorMessage string
// 		expectedPR   *git.PullRequest
// 	}{
// 		"normal": {
// 			body: bytes.NewBuffer([]byte(`{"base_branch": "master", "head_branch": "feat/test"}`)),
// 			expectedPR: &git.PullRequest{
// 				State:  git.PullRequestStateOpen,
// 				Action: git.PullRequestActionOpen,
// 				Author: git.User{
// 					Name: "trigger-test-user-end",
// 				},
// 				Base: git.Base{
// 					Ref: "master",
// 					Sha: git.FakeSha,
// 				},
// 				Head: git.Head{
// 					Ref: "feat/test",
// 					Sha: git.FakeSha,
// 				},
// 			},
// 		},
// 		"decodeErr": {
// 			body:         bytes.NewBuffer([]byte(`{{{{`)),
// 			errorOccurs:  true,
// 			errorMessage: "invalid character '{' looking for beginning of object key string",
// 		},
// 		"defaultBaseBranch": {
// 			body: bytes.NewBuffer([]byte(`{"head_branch": "feat/test"}`)),
// 			expectedPR: &git.PullRequest{
// 				State:  git.PullRequestStateOpen,
// 				Action: git.PullRequestActionOpen,
// 				Author: git.User{
// 					Name: "trigger-test-user-end",
// 				},
// 				Base: git.Base{
// 					Ref: "master",
// 					Sha: git.FakeSha,
// 				},
// 				Head: git.Head{
// 					Ref: "feat/test",
// 					Sha: git.FakeSha,
// 				},
// 			},
// 		},
// 		"noHeadBranch": {
// 			body:         bytes.NewBuffer([]byte(`{}`)),
// 			errorOccurs:  true,
// 			errorMessage: "head_branch must be set",
// 		},
// 	}

// 	for name, c := range tc {
// 		t.Run(name, func(t *testing.T) {
// 			pr, err := buildPullRequestWebhook(c.body, "test-user")
// 			if c.errorOccurs {
// 				require.Error(t, err)
// 				require.Equal(t, c.errorMessage, err.Error())
// 			} else {
// 				require.NoError(t, err)
// 				require.Equal(t, c.expectedPR, pr)
// 			}
// 		})
// 	}
// }

// func Test_buildPushWebhook(t *testing.T) {
// 	tc := map[string]struct {
// 		body io.Reader

// 		errorOccurs  bool
// 		errorMessage string
// 		expectedPush *git.Push
// 	}{
// 		"normal": {
// 			body: bytes.NewBuffer([]byte(`{"branch": "master"}`)),
// 			expectedPush: &git.Push{
// 				Ref: "master",
// 				Sha: "0000000000000000000000000000000000000000",
// 			},
// 		},
// 		"decodeErr": {
// 			body:         bytes.NewBuffer([]byte(`{{{{`)),
// 			errorOccurs:  true,
// 			errorMessage: "invalid character '{' looking for beginning of object key string",
// 		},
// 		"defaultBranch": {
// 			body: bytes.NewBuffer([]byte(`{}`)),
// 			expectedPush: &git.Push{
// 				Ref: "master",
// 				Sha: "0000000000000000000000000000000000000000",
// 			},
// 		},
// 	}

// 	for name, c := range tc {
// 		t.Run(name, func(t *testing.T) {
// 			push, err := buildPushWebhook(c.body)
// 			if c.errorOccurs {
// 				require.Error(t, err)
// 				require.Equal(t, c.errorMessage, err.Error())
// 			} else {
// 				require.NoError(t, err)
// 				require.Equal(t, c.expectedPush, push)
// 			}
// 		})
// 	}
// }
