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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

var testPath = path.Join(os.TempDir(), "cicd-test-kube")
var testKubeConfig = fmt.Sprintf(`apiVersion: v1
Kind: Config
current-context: test-2
clusters:
- name: test-1
  cluster:
    server: https://test-cluster-1
    certificate-authority: %s/ca.crt
- name: test-2
  cluster:
    server: https://test-cluster-2
    certificate-authority: %s/ca2.crt
users:
- name: test-user-1
  user:
    client-certificate: %s/client.crt
    client-key: %s/client.key
- name: test-user-2
  user:
    client-certificate: %s/client2.crt
    client-key: %s/client2.key
contexts:
- name: test-1
  context:
    cluster: test-1
    user: test-user-1
    namespace: test-ns
- name: test-2
  context:
    cluster: test-2
    user: test-user-2
    namespace: test-ns2
`, testPath, testPath, testPath, testPath, testPath, testPath)

func TestLoadKubeConfig(t *testing.T) {
	require.NoError(t, os.MkdirAll(testPath, os.ModePerm))
	defer func() {
		_ = os.RemoveAll(testPath)
	}()
	// Write kube config
	require.NoError(t, ioutil.WriteFile(path.Join(testPath, "config"), []byte(testKubeConfig), os.ModePerm))

	// Write crt, keys
	require.NoError(t, ioutil.WriteFile(path.Join(testPath, "ca.crt"), []byte(""), os.ModePerm))
	require.NoError(t, ioutil.WriteFile(path.Join(testPath, "ca2.crt"), []byte(""), os.ModePerm))
	require.NoError(t, ioutil.WriteFile(path.Join(testPath, "client.crt"), []byte(""), os.ModePerm))
	require.NoError(t, ioutil.WriteFile(path.Join(testPath, "client.key"), []byte(""), os.ModePerm))
	require.NoError(t, ioutil.WriteFile(path.Join(testPath, "client2.crt"), []byte(""), os.ModePerm))
	require.NoError(t, ioutil.WriteFile(path.Join(testPath, "client2.key"), []byte(""), os.ModePerm))

	tc := map[string]struct {
		cfg Configs

		errorOccurs        bool
		errorMessage       string
		expectedKubeConfig *rest.Config
		expectedNamespace  string
	}{
		"basic": {
			cfg: Configs{
				KubeConfig:   path.Join(testPath, "config"),
				CAFile:       path.Join(testPath, "ca.crt"),
				CertFile:     path.Join(testPath, "client.crt"),
				KeyFile:      path.Join(testPath, "client.key"),
				Username:     "test-username",
				Password:     "test-pw",
				AuthInfoName: "test-user-1",
				Namespace:    "test-ns",
				ClusterName:  "test-1",
				Context:      "test-1",
				APIServer:    "https://test-server",
				Insecure:     true,
			},
			expectedNamespace: "test-ns",
			expectedKubeConfig: &rest.Config{
				Host:    "https://test-server",
				APIPath: "/apis",
				ContentConfig: rest.ContentConfig{
					GroupVersion: &schema.GroupVersion{Group: apiGroup, Version: apiVersion},
				},
				Username: "test-username",
				Password: "test-pw",
				TLSClientConfig: rest.TLSClientConfig{
					Insecure: true,
					CertFile: path.Join(testPath, "client.crt"),
					KeyFile:  path.Join(testPath, "client.key"),
					CAFile:   path.Join(testPath, "ca.crt"),
				},
			},
		},
		"token": {
			cfg: Configs{
				KubeConfig:   path.Join(testPath, "config"),
				CAFile:       path.Join(testPath, "ca.crt"),
				CertFile:     path.Join(testPath, "client.crt"),
				KeyFile:      path.Join(testPath, "client.key"),
				BearerToken:  "test-token",
				AuthInfoName: "test-user-1",
				Namespace:    "test-ns",
				ClusterName:  "test-1",
				Context:      "test-1",
				APIServer:    "https://test-server",
				Insecure:     true,
			},
			expectedNamespace: "test-ns",
			expectedKubeConfig: &rest.Config{
				Host:    "https://test-server",
				APIPath: "/apis",
				ContentConfig: rest.ContentConfig{
					GroupVersion: &schema.GroupVersion{Group: apiGroup, Version: apiVersion},
				},
				BearerToken: "test-token",
				TLSClientConfig: rest.TLSClientConfig{
					Insecure: true,
					CertFile: path.Join(testPath, "client.crt"),
					KeyFile:  path.Join(testPath, "client.key"),
					CAFile:   path.Join(testPath, "ca.crt"),
				},
			},
		},
		"cfgErr": {
			cfg:          Configs{KubeConfig: "test/cfg"},
			errorOccurs:  true,
			errorMessage: "test/cfg",
		},
		"nsErr": {
			cfg: Configs{
				KubeConfig:   path.Join(testPath, "config"),
				CAFile:       path.Join(testPath, "ca.crt"),
				CertFile:     path.Join(testPath, "client.crt"),
				KeyFile:      path.Join(testPath, "client.key"),
				BearerToken:  "test-token",
				AuthInfoName: "test-user-1",
				ClusterName:  "test-1",
				Context:      "test-3",
				APIServer:    "https://test-server",
				Insecure:     true,
			},
			errorOccurs:  true,
			errorMessage: "context \"test-3\" does not exist",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			k, ns, err := LoadKubeConfig(&c.cfg)

			if c.errorOccurs {
				require.Error(t, err)
				require.Contains(t, err.Error(), c.errorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedNamespace, ns)

				require.Equal(t, c.expectedKubeConfig.Host, k.Host)
				require.Equal(t, c.expectedKubeConfig.APIPath, k.APIPath)
				require.Equal(t, c.expectedKubeConfig.GroupVersion, k.GroupVersion)
				require.Equal(t, c.expectedKubeConfig.Username, k.Username)
				require.Equal(t, c.expectedKubeConfig.Password, k.Password)
				require.Equal(t, c.expectedKubeConfig.BearerToken, k.BearerToken)
				require.Equal(t, c.expectedKubeConfig.BearerTokenFile, k.BearerTokenFile)
				require.Equal(t, c.expectedKubeConfig.Insecure, k.Insecure)
				require.Equal(t, c.expectedKubeConfig.CertFile, k.CertFile)
				require.Equal(t, c.expectedKubeConfig.KeyFile, k.KeyFile)
				require.Equal(t, c.expectedKubeConfig.CAFile, k.CAFile)
			}
		})
	}
}
