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
