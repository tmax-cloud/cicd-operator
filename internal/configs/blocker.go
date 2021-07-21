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
	if !Initiated {
		Initiated = true
		if len(InitCh) < cap(InitCh) {
			InitCh <- struct{}{}
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
