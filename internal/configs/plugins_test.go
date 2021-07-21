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
