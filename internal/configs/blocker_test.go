package configs

import (
	"github.com/bmizerany/assert"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestApplyBlockerConfigChange(t *testing.T) {
	tc := map[string]controllerTestCase{
		"default": {ConfigMap: &corev1.ConfigMap{
			Data: map[string]string{},
		}, AssertFunc: func(t *testing.T, err error) {
			assert.Equal(t, true, err == nil)

			assert.Equal(t, 1, MergeSyncPeriod)
			assert.Equal(t, "ma/block", MergeBlockLabel)
			assert.Equal(t, "ma/merge-squash", MergeKindSquashLabel)
			assert.Equal(t, "ma/merge-merge", MergeKindMergeLabel)
		}},
		"normal": {ConfigMap: &corev1.ConfigMap{
			Data: map[string]string{
				"mergeSyncPeriod":      "1",
				"mergeBlockLabel":      "test-block",
				"mergeKindSquashLabel": "test-squash",
				"mergeKindMergeLabel":  "test-merge",
			},
		}, AssertFunc: func(t *testing.T, err error) {
			assert.Equal(t, true, err == nil)

			assert.Equal(t, 1, MergeSyncPeriod)
			assert.Equal(t, "test-block", MergeBlockLabel)
			assert.Equal(t, "test-squash", MergeKindSquashLabel)
			assert.Equal(t, "test-merge", MergeKindMergeLabel)
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
