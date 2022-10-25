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
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestApplyBlockerConfigChange(t *testing.T) {
	tc := map[string]controllerTestCase{
		"default": {ConfigMap: &corev1.ConfigMap{
			Data: map[string]string{},
		}, AssertFunc: func(t *testing.T, err error) {
			require.NoError(t, err)

			require.Equal(t, 1, MergeSyncPeriod)
			require.Equal(t, "ci/hold", MergeBlockLabel)
			require.Equal(t, "ci/merge-squash", MergeKindSquashLabel)
			require.Equal(t, "ci/merge-merge", MergeKindMergeLabel)
		}},
		"normal": {ConfigMap: &corev1.ConfigMap{
			Data: map[string]string{
				"mergeSyncPeriod":      "1",
				"mergeBlockLabel":      "test-block",
				"mergeKindSquashLabel": "test-squash",
				"mergeKindMergeLabel":  "test-merge",
			},
		}, AssertFunc: func(t *testing.T, err error) {
			require.NoError(t, err)

			require.Equal(t, 1, MergeSyncPeriod)
			require.Equal(t, "test-block", MergeBlockLabel)
			require.Equal(t, "test-squash", MergeKindSquashLabel)
			require.Equal(t, "test-merge", MergeKindMergeLabel)
		}},
	}

	for name, c := range tc {
		MergeSyncPeriod = 0
		MergeBlockLabel = ""
		MergeKindSquashLabel = ""
		MergeKindMergeLabel = ""
		t.Run(name, func(t *testing.T) {
			err := ApplyBlockerConfigChange(c.ConfigMap)
			c.AssertFunc(t, err)
		})
	}
}
