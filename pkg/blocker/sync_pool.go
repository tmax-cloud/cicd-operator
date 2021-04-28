package blocker

import (
	"context"
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"time"
)

// sync_pool.go contains blocker's methods for synchronizing PR pools.
// It lists pull requests for each IntegrationConfig periodically.
// Then, it checks if the pull requests meet the conditions to be merged. (e.g., branch, label, author, ...)
// (Commit status, merge conflict is not checked here, because those information is not available in list API)
// If all the conditions are met, it is added to a merge pool.

func (b *blocker) loopSyncPRs() {
	b.log.Info("Starting syncPR loop")
	for {
		<-time.After(time.Until(b.lastPoolSync.Add(time.Duration(configs.MergeSyncPeriod) * time.Minute)))
		b.syncPRs()
	}
}

func (b *blocker) syncPRs() {
	log := b.log.WithName("pool")
	log.Info("Synchronizing PR pool")

	// Update lastPoolSync
	b.lastPoolSync = time.Now()

	ics := &cicdv1.IntegrationConfigList{}
	if err := b.client.List(context.Background(), ics); err != nil {
		log.Error(err, "")
		return
	}

	// For each repository - sync PRs and put some into a merge pool
	doneKeys := map[string]struct{}{}
	for _, ic := range ics.Items {
		// Skip if token is nil or merge automation is not activated
		if ic.Spec.Git.Token == nil || ic.Spec.MergeConfig == nil {
			continue
		}
		key := string(genPoolKey(&ic))
		_, done := doneKeys[key]
		if done {
			continue
		}
		doneKeys[key] = struct{}{}

		b.syncOnePool(&ic)
	}

	// Delete redundant pools (i.e., pools for deleted IntegrationConfigs)
	for key := range b.Pools {
		if _, done := doneKeys[string(key)]; !done {
			delete(b.Pools, key)
		}
	}

	// Notify that a sync is done
	if len(b.poolSynced) < cap(b.poolSynced) {
		b.poolSynced <- struct{}{}
	}
}

func (b *blocker) syncOnePool(ic *cicdv1.IntegrationConfig) {
	log := b.log.WithName("pool").WithValues("repo", genPoolKey(ic))

	gitCli, err := utils.GetGitCli(ic, b.client)
	if err != nil {
		b.log.Error(err, "")
		return
	}

	// Init PRPool
	key := genPoolKey(ic)
	if b.Pools[key] == nil {
		b.Pools[key] = NewPRPool(ic.Namespace, ic.Name)
	}

	pool := b.Pools[key]

	prs, err := gitCli.ListPullRequests(true)
	if err != nil {
		b.log.Error(err, "")
		return
	}

	pool.lock.Lock()
	defer pool.lock.Unlock()

	// For each PRs - check status
	prIDs := map[int]struct{}{}
	for _, rawPR := range prs {
		prIDs[rawPR.ID] = struct{}{}

		// Initiate PullRequest object
		pr := pool.PullRequests[rawPR.ID]
		if pr == nil {
			// This should be the one and only place where a PullRequest is created/added to pool.PullRequests
			pr = &PullRequest{
				BlockerStatus:      git.CommitStatusStatePending,
				BlockerDescription: defaultBlockerMessage,
			}
			pool.PullRequests[rawPR.ID] = pr
		}
		pr.PullRequest = rawPR

		// Check conditions (labels, author, branch, conflict)
		isCandidate, addMsg := checkConditionsSimple(ic.Spec.MergeConfig.Query, &rawPR)

		// If it's a re-test from merge pool (i.e., in the merge pool and is in WaitingBatchTest),
		// set it as a candidate and keep it in the merge pool.
		// merger will remove it from the merge pool
		if pool.CurrentBatch != nil && pool.CurrentBatch.Contains(rawPR.ID) {
			isCandidate = true
		}

		// Add to/delete from merge Pool
		if isCandidate {
			// Check if it's in merge pool and if not, add to it
			if pool.MergePool.Search(pr.ID) == nil {
				pool.MergePool.Add(pr)
			}
			// Don't set status here!
			// Sync_status will do the job for the prs in the merge pool
		} else {
			// Delete from merge pool
			pool.MergePool.Delete(pr.ID)

			// Set status
			if pr.BlockerStatus != git.CommitStatusStatePending {
				pr.blockerCacheDirty = true
			}
			pr.BlockerStatus = git.CommitStatusStatePending

			// Append msg
			desc := fmt.Sprintf("%s %s", defaultBlockerMessage, addMsg)
			if pr.BlockerDescription != desc {
				pr.blockerCacheDirty = true
			}
			pr.BlockerDescription = desc
		}

		log.Info(fmt.Sprintf("\t[#%d](%.20s) - merge candidate: %t (%s)", pr.ID, pr.Title, isCandidate, pr.BlockerDescription))
	}

	// Delete redundant PRs (i.e., closed/merged ones)
	for id := range pool.PullRequests {
		_, exist := prIDs[id]
		if !exist {
			pool.MergePool.Delete(id)
			delete(pool.PullRequests, id)
		}
	}
}
