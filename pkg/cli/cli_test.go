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

package cli

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
)

func TestGetClient(t *testing.T) {
	tc := map[string]struct {
		configs Configs

		errorOccurs        bool
		errorMessage       string
		expectedNs         string
		expectedAPIVersion string
		expectedAPIURL     string
	}{
		"normal": {
			configs: Configs{
				APIServer:   "https://test.cluster.com",
				Namespace:   "default",
				BearerToken: "test-token",
			},
			expectedNs:         "default",
			expectedAPIVersion: "cicdapi.tmax.io/v1",
			expectedAPIURL:     "https://test.cluster.com/apis/cicdapi.tmax.io/v1",
		},
		"loadErr": {
			configs: Configs{
				KubeConfig: "test-cfg",
			},
			errorOccurs:  true,
			errorMessage: "test-cfg",
		},
		"restErr": {
			configs: Configs{
				APIServer: "tcp:/dd/test.cluster.com",
			},
			errorOccurs:  true,
			errorMessage: "host must be a URL or a host:port pair: \"tcp:///dd/test.cluster.com\"",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			cli, ns, err := GetClient(&c.configs)
			if c.errorOccurs {
				require.Error(t, err)
				require.Contains(t, err.Error(), c.errorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedNs, ns)
				require.Equal(t, c.expectedAPIVersion, cli.APIVersion().String())
				require.Equal(t, c.expectedAPIURL, cli.Get().URL().String())
			}
		})
	}
}

func TestExecAndHandleError(t *testing.T) {
	testSrv := testAPIServer()
	uri, err := url.Parse(testSrv.URL)
	require.NoError(t, err)

	tc := map[string]struct {
		req *rest.Request
		fn  func([]byte) error

		errorOccurs  bool
		errorMessage string
	}{
		"normal": {
			req: rest.NewRequestWithClient(uri, "", rest.ClientContentConfig{}, http.DefaultClient),
			fn: func(bytes []byte) error {
				return nil
			},
		},
		"noFn": {
			req: rest.NewRequestWithClient(uri, "", rest.ClientContentConfig{}, http.DefaultClient),
		},
		"reqErr": {
			req:          rest.NewRequestWithClient(uri, "/errJson", rest.ClientContentConfig{}, http.DefaultClient),
			errorOccurs:  true,
			errorMessage: "some error",
		},
		"respMarshalErr": {
			req:          rest.NewRequestWithClient(uri, "/err", rest.ClientContentConfig{}, http.DefaultClient),
			errorOccurs:  true,
			errorMessage: "unexpected end of JSON input",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			err := ExecAndHandleError(c.req, c.fn)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func testAPIServer() *httptest.Server {
	router := mux.NewRouter()
	router.HandleFunc("/", func(_ http.ResponseWriter, _ *http.Request) {})
	router.HandleFunc("/err", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	router.HandleFunc("/errJson", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message": "some error"}`))
	})
	return httptest.NewServer(router)
}
