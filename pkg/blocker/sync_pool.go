package blocker

import (
	"context"
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
)

// sync_pool.go contains blocker's methods for synchronizing PR pools.
// It lists pull requests for each IntegrationConfig periodically.
// Then, it checks if the pull requests meet the conditions to be merged. (e.g., branch, label, author, ...)
// (Commit status, merge conflict is not checked here, because those information is not available in list API)
// If all the conditions are met, it is added to a merge pool.

func (b *blocker) loopSyncPRs() {
	// TODO - call b.syncPRs() every SyncDuration
	//wait := time.After(time.Until(lastSyncStart.Add(SyncDuration)))
	//for {
	//	<- wait
	//	b.syncPRs()
	//}
}

func (b *blocker) syncPRs() {
	ics := &cicdv1.IntegrationConfigList{}
	if err := b.client.List(context.Background(), ics); err != nil {
		b.log.Error(err, "")
		return
	}

	// For each IntegrationConfig
	var icKeys []string
	for _, ic := range ics.Items {
		// Skip if token is nil or merge automation is not activated
		if ic.Spec.Git.Token == nil || ic.Spec.MergeConfig == nil {
			continue
		}
		icKeys = append(icKeys, string(poolKey(&ic)))

		b.syncOnePool(&ic)
	}

	// Delete redundant pools (i.e., pools for deleted IntegrationConfigs)
	for key := range b.Pools {
		if !containsString(string(key), icKeys) {
			delete(b.Pools, key)
		}
	}
}

func (b *blocker) syncOnePool(ic *cicdv1.IntegrationConfig) {
	log := b.log.WithValues("namespace", ic.Namespace, "name", ic.Name)
	log.Info("Synchronizing PR pool")

	gitCli, err := utils.GetGitCli(ic, b.client)
	if err != nil {
		b.log.Error(err, "")
		return
	}

	// Init PRPool
	key := poolKey(ic)
	if b.Pools[key] == nil {
		b.Pools[key] = NewPRPool()
	}

	pool := b.Pools[key]

	prs, err := gitCli.ListPullRequests(true)
	if err != nil {
		b.log.Error(err, "")
		return
	}

	pool.lock.Lock()
	defer pool.lock.Unlock()

	var prIDs []int
	for _, rawPR := range prs {
		prIDs = append(prIDs, rawPR.ID)

		// Initiate PullRequest object
		pr := pool.PullRequests[rawPR.ID]
		if pr == nil {
			// This should be the one and only place where a PullRequest is created/added to pool.PullRequests
			pr = &PullRequest{
				BlockerStatus:      cicdv1.CommitStatusStatePending,
				BlockerDescription: defaultBlockerMessage,
			}
			pool.PullRequests[rawPR.ID] = pr
		}
		pr.PullRequest = rawPR

		// Check conditions (labels, author, branch, conflict)
		isCandidate, addMsg := checkConditionsSimple(ic.Spec.MergeConfig.Query, &rawPR)

		// Add to/delete from merge Pool
		if isCandidate {
			// Check if it's in merge pool and if not, add to it
			if pool.MergePool.Search(pr.ID) == nil {
				pool.MergePool.Add(pr)
			}

			// Set default message
			pr.BlockerDescription = defaultBlockerMessage
		} else {
			// Delete from merge pool
			pool.MergePool.Delete(pr.ID)

			// Append msg
			pr.BlockerDescription = fmt.Sprintf("%s %s", defaultBlockerMessage, addMsg)
		}

		log.Info(fmt.Sprintf("[#%d](%.20s) - %s/%s", pr.ID, pr.Title, pr.BlockerStatus, pr.BlockerDescription))
	}

	// Delete redundant PRs (i.e., closed/merged ones)
	for id := range pool.PullRequests {
		if !containsInt(id, prIDs) {
			pool.MergePool.Delete(id)
			delete(pool.PullRequests, id)
		}
	}
}
