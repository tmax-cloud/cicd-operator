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

	// Notify that a sync is done
	if len(b.statusSynced) < cap(b.statusSynced) {
		b.statusSynced <- struct{}{}
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
			// Fetch PR's status, commit statuses
			if err := b.reflectPRStatus(pr, gitCli); err != nil {
				log.Error(err, "")
				continue
			}
			newStatusB, removeFromMergePool, newDescription := checkConditionsFull(ic.Spec.MergeConfig.Query, pr)

			var newStatus git.CommitStatusState
			if newStatusB {
				newStatus = git.CommitStatusStateSuccess
				newDescription = "In merge pool."
			} else {
				newStatus = git.CommitStatusStatePending
			}

			// Remove from merge pool if simple test fails
			// But, if the PR is being re-tested by merger, keep it in the merge pool
			if removeFromMergePool && (pool.CurrentBatch == nil || !pool.CurrentBatch.Contains(prID)) {
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

func (b *blocker) reflectPRStatus(pull *PullRequest, gitCli git.Client) error {
	// GET PullRequest
	pr, err := gitCli.GetPullRequest(pull.ID)
	if err != nil {
		return err
	}
	pull.PullRequest = *pr

	// GET PR statuses
	checksSlice, err := gitCli.ListCommitStatuses(pr.Head.Sha)
	if err != nil {
		return err
	}
	pull.Statuses = map[string]git.CommitStatus{}
	for _, c := range checksSlice {
		pull.Statuses[c.Context] = c
	}
	return nil
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
			if err := gitCli.SetCommitStatus(pr.Head.Sha, git.CommitStatus{Context: blockerContext, State: pr.BlockerStatus, Description: pr.BlockerDescription, TargetURL: blockerURL}); err != nil {
				log.Error(err, "")
				continue
			}
		}
	}
}
