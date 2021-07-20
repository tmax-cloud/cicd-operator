package configs

import corev1 "k8s.io/api/core/v1"

// ConfigMapNamePluginConfig is a name of plugin config map
const (
	ConfigMapNamePluginConfig = "plugin-config"
)

// ApplyPluginConfigChange is a configmap handler for plugin-config configmap
func ApplyPluginConfigChange(cm *corev1.ConfigMap) error {
	getVars(cm.Data, map[string]operatorConfig{
		"sizeS":   {Type: cfgTypeInt, IntVal: &PluginSizeS, IntDefault: 10},
		"sizeM":   {Type: cfgTypeInt, IntVal: &PluginSizeM, IntDefault: 30},
		"sizeL":   {Type: cfgTypeInt, IntVal: &PluginSizeL, IntDefault: 100},
		"sizeXL":  {Type: cfgTypeInt, IntVal: &PluginSizeXL, IntDefault: 500},
		"sizeXXL": {Type: cfgTypeInt, IntVal: &PluginSizeXXL, IntDefault: 1000},
	})
	return nil
}

// Configs for Size plugin
var (
	PluginSizeS   = 10
	PluginSizeM   = 30
	PluginSizeL   = 100
	PluginSizeXL  = 500
	PluginSizeXXL = 1000
)
