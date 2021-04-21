package blocker

import (
	"context"
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
)

// sync_pool.go includes methods for synchronizing PR's commit status/merge conflicts status
// Also, it updates commit status for all open PRs (not only those in merge pool) if needed

func (b *blocker) loopSyncMergePoolStatus() {
	// Call sync pool status method whenever a PR pool sync is done
	for range b.poolSynced {
		b.syncMergePoolStatus()
	}
}

func (b *blocker) syncMergePoolStatus() {
	log := b.log.WithName("status")
	log.Info("Synchronizing merge pool status")

	for _, pool := range b.Pools {
		log := b.log.WithName("status").WithValues("repo", pool.NamespacedName)

		// Get IC
		ic := &cicdv1.IntegrationConfig{}
		if err := b.client.Get(context.Background(), pool.NamespacedName, ic); err != nil {
			log.Error(err, "")
			continue
		}
		if ic.Spec.MergeConfig == nil {
			continue
		}

		gitCli, err := utils.GetGitCli(ic, b.client)
		if err != nil {
			log.Error(err, "")
			continue
		}

		// Get PRs' status for each repository
		b.syncOneMergePoolStatus(pool, ic, gitCli)

		// Set PRs' commit status for each repository
		b.reportCommitStatus(pool, ic, gitCli)
	}
}

func (b *blocker) syncOneMergePoolStatus(pool *PRPool, ic *cicdv1.IntegrationConfig, gitCli git.Client) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	log := b.log.WithName("status").WithValues("repo", genPoolKey(ic))

	// Loop merge pool per blocker status (pending, success)
	for oldStatus := range pool.MergePool {
		// For each PR
		for prID, pr := range pool.MergePool[oldStatus] {
			newStatusB, removeFromMergePool, newDescription, err := checkConditionsFull(ic.Spec.MergeConfig.Query, pr.ID, gitCli)
			if err != nil {
				log.Error(err, "")
				continue
			}

			var newStatus git.CommitStatusState
			if newStatusB {
				newStatus = git.CommitStatusStateSuccess
				newDescription = "In merge pool."
			} else {
				newStatus = git.CommitStatusStatePending
			}

			// Remove from merge pool if simple test fails
			if removeFromMergePool {
				delete(pool.MergePool[oldStatus], prID)
			}

			// Move PR status in the pool
			if newStatus != oldStatus {
				delete(pool.MergePool[oldStatus], prID)
				pool.MergePool[newStatus][prID] = pr
			}

			// Update status cache
			if newStatus != pr.BlockerStatus {
				pr.BlockerStatus = newStatus
				pr.blockerCacheDirty = true
			}
			if newDescription != pr.BlockerDescription {
				pr.BlockerDescription = newDescription
				pr.blockerCacheDirty = true
			}

			log.Info(fmt.Sprintf("\t[#%d](%.20s) - %s/%s", pr.ID, pr.Title, pr.BlockerStatus, pr.BlockerDescription))
		}
	}
}

func (b *blocker) reportCommitStatus(pool *PRPool, ic *cicdv1.IntegrationConfig, gitCli git.Client) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	log := b.log.WithName("status").WithValues("repo", genPoolKey(ic))

	// Update commit status
	for _, pr := range pool.PullRequests {
		// Update commit status
		if pr.blockerCacheDirty {
			pr.blockerCacheDirty = false
			blockerURL := "" // TODO
			log.Info(fmt.Sprintf("Setting commit status %s:%s:%s to %s's %s", blockerContext, pr.BlockerStatus, pr.BlockerDescription, pool.NamespacedName.String(), pr.Head.Sha))
			if err := gitCli.SetCommitStatus(pr.Head.Sha, blockerContext, pr.BlockerStatus, pr.BlockerDescription, blockerURL); err != nil {
				log.Error(err, "")
				continue
			}
		}
	}
}
