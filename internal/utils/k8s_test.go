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

package utils

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNamespace(t *testing.T) {
	tc := map[string]struct {
		path string
		data []byte
		env  string

		expectedNs string
	}{
		"inCluster": {
			path:       path.Join(os.TempDir(), "test-cicd-operator-ns-file-in-cluster"),
			data:       []byte("cicd-system3"),
			expectedNs: "cicd-system3",
		},
		"env": {
			path:       path.Join(os.TempDir(), "test-cicd-operator-ns-file-env"),
			env:        "cicd-system2",
			expectedNs: "cicd-system2",
		},
		"default": {
			path:       path.Join(os.TempDir(), "test-cicd-operator-ns-file-default"),
			expectedNs: "cicd-system",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			nsFilePath = c.path
			if c.path != "" && len(c.data) > 0 {
				require.NoError(t, os.MkdirAll(path.Dir(c.path), os.ModePerm))
				require.NoError(t, ioutil.WriteFile(c.path, c.data, os.ModePerm))
				defer func() {
					_ = os.RemoveAll(c.path)
				}()
			}

			if c.env != "" {
				require.NoError(t, os.Setenv(nsEnv, c.env))
				defer func() {
					_ = os.Unsetenv(nsEnv)
				}()
			}

			ns := Namespace()

			require.Equal(t, c.expectedNs, ns)
		})
	}
}
