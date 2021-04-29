# Configuring Blocker

This guide shows how to configure the blocker. Contents are as follows.
- [`maxPipelineRun`](#maxpipelinerun)
- [`ingressClass`](#ingressclass)
- [`externalHostName`](#externalhostname)

You can check and update the configuration values from the ConfigMap `blocker-config` in namespace `cicd-system`.
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: blocker-config
  namespace: cicd-system
data:
  mergeSyncPeriod: "1" # in minute
  mergeBlockLabel: "ci/hold"
  mergeKindSquashLabel: "ci/merge-squash"
  mergeKindMergeLabel: "ci/merge-merge"
```

### `mergeSyncPeriod`
Period (in minute) of synchronizing a merge pool. If it's set to `1`, we check if pull requests are ready to be merged and merge them every 1 minute.
> Default: 1 (m)

### `mergeBlockLabel`
Label to block the pull request from being merged. If you put the label to a pull request, it's not merged even if its' merge conditions are all satisfied.

### `mergeKindSquashLabel`
Label to make the pull request to be merged with `squash` method. If you put the label to a pull request, it is merged with `squash` method, no matter what method is configured to MergeConfig.

### `mergeKindMergeLabel`
Label to make the pull request to be merged with `merge` method. If you put the label to a pull request, it is merged with `merge` method, no matter what method is configured to MergeConfig.
