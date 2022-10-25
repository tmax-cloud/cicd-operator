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
	"context"
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
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_handler_webhookURLHandler(t *testing.T) {
	tc := map[string]struct {
		ic   *cicdv1.IntegrationConfig
		vars map[string]string

		expectedCode    int
		expectedMessage string
	}{
		"normal": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeGitHub,
						Repository: "test-repo",
					},
				},
				Status: cicdv1.IntegrationConfigStatus{Secrets: "test-secret"},
			},
			vars: map[string]string{
				"namespace": "test-ns",
				"icName":    "test-ic",
			},
			expectedCode:    200,
			expectedMessage: "{\"url\":\"http:///webhook/test-ns/test-ic\",\"secret\":\"test-secret\"}",
		},
		"noVars": {
			ic: &cicdv1.IntegrationConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-ic", Namespace: "test-ns"},
				Spec: cicdv1.IntegrationConfigSpec{
					Git: cicdv1.GitConfig{
						Type:       cicdv1.GitTypeGitHub,
						Repository: "test-repo",
					},
				},
				Status: cicdv1.IntegrationConfigStatus{Secrets: "test-secret"},
			},
			expectedCode:    400,
			expectedMessage: "url is malformed",
		},
		"icGetErr": {
			vars: map[string]string{
				"namespace": "test-ns",
				"icName":    "test-ic",
			},
			expectedCode:    500,
			expectedMessage: "cannot get IntegrationConfig test-ns/test-ic",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			s := runtime.NewScheme()
			require.NoError(t, cicdv1.AddToScheme(s))

			fakeCli := fake.NewClientBuilder().WithScheme(s).Build()
			if c.ic != nil {
				require.NoError(t, fakeCli.Create(context.Background(), c.ic))
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/", nil)
			req = mux.SetURLVars(req, c.vars)

			handler := &handler{log: &test.FakeLogger{}, k8sClient: fakeCli}
			handler.webhookURLHandler(w, req)

			require.Equal(t, c.expectedCode, w.Result().StatusCode)
			b, err := ioutil.ReadAll(w.Result().Body)
			require.NoError(t, err)
			require.Contains(t, string(b), c.expectedMessage)
		})
	}
}
