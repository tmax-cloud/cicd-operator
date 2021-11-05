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

package apiserver

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/tmax-cloud/cicd-operator/internal/test"
	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	testing2 "k8s.io/client-go/testing"
)

func TestNewAuthorizer(t *testing.T) {
	_ = NewAuthorizer(fake.NewSimpleClientset().AuthorizationV1(), "", "", "")
}

func Test_authorizer_Authorize(t *testing.T) {
	tc := map[string]struct {
		path   string
		tls    *tls.ConnectionState
		header http.Header

		errorOccurs     bool
		errorMessage    string
		expectedCode    int
		expectedMessage string
	}{
		"normal": {
			path: "/apis/test.api.group/v1/namespaces/test-ns/testresources/test-resource/subresource",
			tls:  &tls.ConnectionState{PeerCertificates: []*x509.Certificate{{Version: 32}}},
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			expectedCode:    200,
			expectedMessage: "",
		},
		"noTLS": {
			path: "/apis/test.api.group/v1/namespaces/test-ns/testresources/test-resource/subresource",
			header: map[string][]string{
				"X-Remote-User":  {"test-user"},
				"X-Remote-Group": {"test-group"},
			},
			expectedCode:    401,
			expectedMessage: "{\"message\":\"is not https or there is no peer certificate\"}",
		},
		"reviewErr": {
			path: "/apis/test.api.group/v1/namespaces/test-ns/testresources/test-resource/subresource",
			tls:  &tls.ConnectionState{PeerCertificates: []*x509.Certificate{{Version: 32}}},
			header: map[string][]string{
				"X-Remote-Group": {"test-group"},
			},
			expectedCode:    403,
			expectedMessage: "{\"message\":\"no header X-Remote-User\"}",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			fakeSet := fake.NewSimpleClientset()
			fakeSet.ReactionChain = append([]testing2.Reactor{&fakeReactor{}}, fakeSet.ReactionChain...)
			fakeCli := fakeSet.AuthorizationV1()

			a := &authorizer{
				log:     &test.FakeLogger{},
				AuthCli: fakeCli,
			}

			router := mux.NewRouter()
			router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.URL.Path = c.path
			req.TLS = c.tls
			req.Header = c.header
			a.Authorize(router).ServeHTTP(w, req)

			b, err := ioutil.ReadAll(w.Result().Body)
			require.NoError(t, err)
			require.Equal(t, c.expectedMessage, string(b))

			require.Equal(t, c.expectedCode, w.Result().StatusCode)
		})
	}
}

func Test_authorizer_reviewAccess(t *testing.T) {
	tc := map[string]struct {
		req *http.Request

		errorOccurs  bool
		errorMessage string
	}{
		"allowed": {
			req: &http.Request{
				URL: &url.URL{
					Path: "/apis/test.api.group/v1/namespaces/test-ns/testresources/test-resource/subresource",
				},
				Header: map[string][]string{
					"X-Remote-User":  {"test-user"},
					"X-Remote-Group": {"test-group"},
				},
			},
		},
		"notAllowed": {
			req: &http.Request{
				URL: &url.URL{
					Path: "/apis/test.api.group/v1/namespaces/test-ns/testresources/test-resource/subresource",
				},
				Header: map[string][]string{
					"X-Remote-User":  {"test-user-not-allowed"},
					"X-Remote-Group": {"test-group"},
				},
			},
			errorOccurs:  true,
			errorMessage: "user test-user-not-allowed is not allowed",
		},
		"noUserName": {
			req: &http.Request{
				URL: &url.URL{
					Path: "/apis/test.api.group/v1/namespaces/test-ns/testresources/test-resource/subresource",
				},
				Header: map[string][]string{},
			},
			errorOccurs:  true,
			errorMessage: "no header X-Remote-User",
		},
		"noGroup": {
			req: &http.Request{
				URL: &url.URL{
					Path: "/apis/test.api.group/v1/namespaces/test-ns/testresources/test-resource/subresource",
				},
				Header: map[string][]string{
					"X-Remote-User": {"test-user"},
				},
			},
			errorOccurs:  true,
			errorMessage: "no header X-Remote-Group",
		},
		"urlMalformed": {
			req: &http.Request{
				URL: &url.URL{
					Path: "/apis/test.api.group/v1/namespaces/test-ns/testresources/test-resource",
				},
				Header: map[string][]string{
					"X-Remote-User":  {"test-user"},
					"X-Remote-Group": {"test-group"},
				},
			},
			errorOccurs:  true,
			errorMessage: "URL should be in form of '/apis/<ApiGroup>/<ApiVersion>/namespaces/<Namespace>/<Resource>/<ResourceName>/<SubResource>'",
		},
		"reviewError": {
			req: &http.Request{
				URL: &url.URL{
					Path: "/apis/test.api.group/v1/namespaces/test-ns/testresources/test-resource/subresource",
				},
				Header: map[string][]string{
					"X-Remote-User":  {"test-user-error"},
					"X-Remote-Group": {"test-group"},
				},
			},
			errorOccurs:  true,
			errorMessage: "test-user-error returns error",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			fakeSet := fake.NewSimpleClientset()
			fakeSet.ReactionChain = append([]testing2.Reactor{&fakeReactor{}}, fakeSet.ReactionChain...)
			fakeCli := fakeSet.AuthorizationV1()

			a := &authorizer{
				log:     &test.FakeLogger{},
				AuthCli: fakeCli,
			}

			err := a.reviewAccess(c.req)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type fakeReactor struct{}

func (f *fakeReactor) Handles(action testing2.Action) bool {
	return action.GetVerb() == "create" && action.GetResource().Resource == "subjectaccessreviews"
}

func (f *fakeReactor) React(action testing2.Action) (bool, runtime.Object, error) {
	cr, ok := action.(testing2.CreateActionImpl)
	if !ok {
		return false, nil, nil
	}
	r, ok := cr.Object.(*authorizationv1.SubjectAccessReview)
	if !ok {
		return false, nil, nil
	}

	r.Status = authorizationv1.SubjectAccessReviewStatus{}
	if r.Spec.User == "test-user" {
		r.Status.Allowed = true
	} else if r.Spec.User == "test-user-error" {
		return true, r, fmt.Errorf("test-user-error returns error")
	} else {
		r.Status.Allowed = false
		r.Status.Reason = fmt.Sprintf("user %s is not allowed", r.Spec.User)
	}

	return true, r, nil
}
