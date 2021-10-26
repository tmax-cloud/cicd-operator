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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/tmax-cloud/cicd-operator/internal/wrapper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	apiregv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestNew(t *testing.T) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}
	fakeCli := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	cfg := &rest.Config{Host: "https://test"}
	c, err := cache.New(cfg, cache.Options{Mapper: &meta.DefaultRESTMapper{}})
	require.NoError(t, err)

	srv := New(fakeCli, cfg, c)
	require.NotNil(t, srv.wrapper)
	require.NotNil(t, srv.wrapper.Router())
	require.Equal(t, fakeCli, srv.client)
	require.Equal(t, c, srv.cache)
	require.NotNil(t, srv.apisHandler)
}

func Test_serverStart(t *testing.T) {
	if _, exist := os.LookupEnv("CI"); !exist {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	}

	utilruntime.Must(apiregv1.AddToScheme(scheme.Scheme))

	apiSvc := &apiregv1.APIService{
		ObjectMeta: metav1.ObjectMeta{
			Name: APIServiceName,
		},
		Spec: apiregv1.APIServiceSpec{},
	}

	extAuth := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "extension-apiserver-authentication",
			Namespace: "kube-system",
		},
		Data: map[string]string{
			"requestheader-client-ca-file": `-----BEGIN CERTIFICATE-----
MIIBaTCCAQ6gAwIBAgIBADAKBggqhkjOPQQDAjArMSkwJwYDVQQDDCBrM3MtcmVx
dWVzdC1oZWFkZXItY2FAMTU5NTU0OTQ0NjAeFw0yMDA3MjQwMDEwNDZaFw0zMDA3
MjIwMDEwNDZaMCsxKTAnBgNVBAMMIGszcy1yZXF1ZXN0LWhlYWRlci1jYUAxNTk1
NTQ5NDQ2MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE85FjW9alEGKy8Jcavk2b
+hvgPV6XxgXnuz0G9RMxLsKu9SXVnaMRc2L9nXTnYOuz5b2FlnTdCWp7WTt35YVQ
VKMjMCEwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB/wQFMAMBAf8wCgYIKoZIzj0E
AwIDSQAwRgIhALCUqk9KPgxhXs+ka5oBnMVgP/xDd33WooGChkXCdLXXAiEA9YQX
rcFz1g2uGUgBe3mBBDID0wosv/64zWA1x4uuwuM=
-----END CERTIFICATE-----`,
		},
	}

	fCli := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(apiSvc, extAuth).Build()
	fCache := &informertest.FakeInformers{}

	srv := &server{
		cache:   fCache,
		client:  fCli,
		wrapper: wrapper.New("/", nil, nil),
	}
	srv.wrapper.SetRouter(mux.NewRouter())

	go srv.Start()
}

func Test_server_rootHandler(t *testing.T) {
	srv := &server{}
	srv.wrapper = wrapper.New("/", nil, nil)
	srv.wrapper.SetRouter(mux.NewRouter())
	require.NoError(t, srv.wrapper.Add(wrapper.New("/apis", []string{http.MethodGet}, func(w http.ResponseWriter, req *http.Request) {})))

	re := httptest.NewRecorder()
	srv.rootHandler(re, httptest.NewRequest(http.MethodPut, "/", nil))

	b, err := ioutil.ReadAll(re.Body)
	require.NoError(t, err)
	require.Equal(t, []byte("{\"paths\":[\"/apis\"]}"), b)
}

func Test_addPath(t *testing.T) {
	wr := wrapper.New("/api", []string{http.MethodGet}, func(w http.ResponseWriter, req *http.Request) {})
	wr.SetRouter(mux.NewRouter())
	require.NoError(t, wr.Add(wrapper.New("/v1", []string{http.MethodGet}, func(w http.ResponseWriter, req *http.Request) {})))

	paths := []string{"/"}
	addPath(&paths, wr)

	require.Len(t, paths, 3)
	require.Equal(t, "/", paths[0])
	require.Equal(t, "/api", paths[1])
	require.Equal(t, "/api/v1", paths[2])
}
