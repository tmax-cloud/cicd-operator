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

package configs

import (
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestApplyPluginConfigChange(t *testing.T) {
	tc := map[string]controllerTestCase{
		"default": {ConfigMap: &corev1.ConfigMap{
			Data: map[string]string{},
		}, AssertFunc: func(t *testing.T, err error) {
			require.NoError(t, err)
			require.Equal(t, 10, PluginSizeS)
			require.Equal(t, 30, PluginSizeM)
			require.Equal(t, 100, PluginSizeL)
			require.Equal(t, 500, PluginSizeXL)
			require.Equal(t, 1000, PluginSizeXXL)
		}},
		"normal": {ConfigMap: &corev1.ConfigMap{
			Data: map[string]string{
				"sizeS":   "100",
				"sizeM":   "200",
				"sizeL":   "300",
				"sizeXL":  "500",
				"sizeXXL": "800",
			},
		}, AssertFunc: func(t *testing.T, err error) {
			require.NoError(t, err)
			require.Equal(t, 100, PluginSizeS)
			require.Equal(t, 200, PluginSizeM)
			require.Equal(t, 300, PluginSizeL)
			require.Equal(t, 500, PluginSizeXL)
			require.Equal(t, 800, PluginSizeXXL)
		}},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			err := ApplyPluginConfigChange(c.ConfigMap)
			c.AssertFunc(t, err)
		})
	}
}
