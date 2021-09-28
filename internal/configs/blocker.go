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
	corev1 "k8s.io/api/core/v1"
)

// ConfigMapNameBlockerConfig is a name of a configmap for a blocker
const (
	ConfigMapNameBlockerConfig = "blocker-config"
)

// ApplyBlockerConfigChange is a configmap handler for blocker-config configmap
func ApplyBlockerConfigChange(cm *corev1.ConfigMap) error {
	getVars(cm.Data, map[string]operatorConfig{
		"mergeSyncPeriod":      {Type: cfgTypeInt, IntVal: &MergeSyncPeriod, IntDefault: 1},                               // Merge automation sync period
		"mergeBlockLabel":      {Type: cfgTypeString, StringVal: &MergeBlockLabel, StringDefault: "ci/hold"},              // Merge automation block label
		"mergeKindSquashLabel": {Type: cfgTypeString, StringVal: &MergeKindSquashLabel, StringDefault: "ci/merge-squash"}, // Merge kind squash label
		"mergeKindMergeLabel":  {Type: cfgTypeString, StringVal: &MergeKindMergeLabel, StringDefault: "ci/merge-merge"},   // Merge kind squash label
	})

	// Init
	if !BlockerInitiated {
		BlockerInitiated = true
		if len(BlockerInitCh) < cap(BlockerInitCh) {
			BlockerInitCh <- struct{}{}
		}
	}

	return nil
}

// Merge Automation Configs
var (
	// MergeSyncPeriod is a PR sync period in minute
	MergeSyncPeriod int

	// MergeBlockLabel is a label name which blocks a PR to be merged
	MergeBlockLabel string

	// MergeKindSquashLabel is a label to make a PR to be merged by 'squash'
	MergeKindSquashLabel string

	// MergeKindMergeLabel is a label to make a PR to be merged by 'merge'
	MergeKindMergeLabel string
)
