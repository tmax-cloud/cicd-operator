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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/gorilla/mux"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/apiserver"
	"github.com/tmax-cloud/cicd-operator/pkg/chatops"
	"github.com/tmax-cloud/cicd-operator/pkg/dispatcher"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"github.com/tmax-cloud/cicd-operator/pkg/server"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type assertFunc func(*cicdv1.IntegrationJob)

func Test_handler_runPreHandler(t *testing.T) {
	h, err := initTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	// TEST 1
	testReq(t, h, cicdv1.IntegrationConfigAPIReqRunPreBody{HeadBranch: "newnew"}, "runpre", func(job *cicdv1.IntegrationJob) {
		assert.Equal(t, true, job.Spec.Refs.Pulls != nil, "refs.pr is not nil")
		assert.Equal(t, "newnew", job.Spec.Refs.Pulls[0].Ref.String(), "pr.ref")
	})

	// TEST 2
	testReq(t, h, cicdv1.IntegrationConfigAPIReqRunPreBody{HeadBranch: "newnew", BaseBranch: "new"}, "runpre", func(job *cicdv1.IntegrationJob) {
		assert.Equal(t, "new", job.Spec.Refs.Base.Ref.String(), "base.ref")
		assert.Equal(t, true, job.Spec.Refs.Pulls != nil, "refs.pr is not nil")
		assert.Equal(t, "newnew", job.Spec.Refs.Pulls[0].Ref.String(), "pr.ref")
	})
}

func Test_handler_runPostHandler(t *testing.T) {
	h, err := initTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	// TEST 1
	testReq(t, h, cicdv1.IntegrationConfigAPIReqRunPostBody{}, "runpost", func(job *cicdv1.IntegrationJob) {
		assert.Equal(t, "master", job.Spec.Refs.Base.Ref.String(), "base ref")
		assert.Equal(t, true, job.Spec.Refs.Pulls == nil, "ref.pr is nil")
	})

	// TEST 2
	testReq(t, h, cicdv1.IntegrationConfigAPIReqRunPostBody{Branch: "new"}, "runpost", func(job *cicdv1.IntegrationJob) {
		assert.Equal(t, "new", job.Spec.Refs.Base.Ref.String(), "base ref")
		assert.Equal(t, true, job.Spec.Refs.Pulls == nil, "ref.pr is nil")
	})
}

func initTestEnv() (*handler, error) {
	s := runtime.NewScheme()
	utilruntime.Must(cicdv1.AddToScheme(s))

	ic := &cicdv1.IntegrationConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: cicdv1.IntegrationConfigSpec{
			Git: cicdv1.GitConfig{
				Type:       "github",
				Repository: "tmax-cloud/cicd-operator",
			},
			Jobs: cicdv1.IntegrationConfigJobs{
				PreSubmit: cicdv1.Jobs{{
					Container: corev1.Container{
						Name:  "pre-test",
						Image: "test-img",
					},
				}},
				PostSubmit: cicdv1.Jobs{{
					Container: corev1.Container{
						Name:  "post-test",
						Image: "test-img",
					},
				}},
			},
		},
	}
	fakeCli := fake.NewFakeClientWithScheme(s, ic)

	h := &handler{
		k8sClient: fakeCli,
		log:       logf.Log,
	}

	server.AddPlugin([]git.EventType{git.EventTypePullRequest, git.EventTypePush}, &dispatcher.Dispatcher{Client: h.k8sClient})
	server.AddPlugin([]git.EventType{git.EventTypeIssueComment}, chatops.New(h.k8sClient))

	return h, nil
}

func testReq(t *testing.T, h *handler, body interface{}, subresource string, assertFunc assertFunc) {
	req, w := newTestReq(body, subresource)
	switch subresource {
	case "runpost":
		h.runPostHandler(w, req)
	case "runpre":
		h.runPreHandler(w, req)
	default:
		t.Fatal("subresource " + subresource + " is not supported")
	}
	list := &cicdv1.IntegrationJobList{}
	if err := h.k8sClient.List(context.Background(), list); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, w.Code, "response code")
	assert.Equal(t, 1, len(list.Items), "ij length")

	assert.Equal(t, "tmax-cloud/cicd-operator", list.Items[0].Spec.Refs.Repository, "repository")
	assert.Equal(t, "https://github.com/tmax-cloud/cicd-operator", list.Items[0].Spec.Refs.Base.Link, "refs.base link")
	assert.Equal(t, "trigger-system_serviceaccount_default_admin-end", list.Items[0].Spec.Refs.Sender.Name, "sender name")
	assert.Equal(t, "https://github.com/tmax-cloud/cicd-operator", list.Items[0].Spec.Refs.Link, "refs link")

	if assertFunc != nil {
		assertFunc(&list.Items[0])
	}

	if err := h.k8sClient.Delete(context.Background(), &list.Items[0]); err != nil {
		t.Fatal(err)
	}
}

func newTestReq(body interface{}, subresource string) (*http.Request, *httptest.ResponseRecorder) {
	var buf bytes.Buffer
	b, _ := json.Marshal(body)
	buf.Write(b)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/apis/cicdapi.tmax.io/v1/namespaces/default/integrationconfigs/test/"+subresource, &buf)
	req.Header.Add("X-Remote-User", "system:serviceaccount:default:admin")
	req = mux.SetURLVars(req, map[string]string{
		apiserver.NamespaceParamKey: "default",
		icParamKey:                  "test",
	})
	w := httptest.NewRecorder()
	return req, w
}
