package blocker

import (
	"context"
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/dispatcher"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"github.com/tmax-cloud/cicd-operator/pkg/pipelinemanager"
	"k8s.io/apimachinery/pkg/types"
	"sort"
)

const (
	maxBatchSize = 1 // TODO - currently, IntegrationJob does not support batch test
)

func (b *blocker) loopMerge() {
	// Call retestAndMerge method whenever a status sync is done
	for range b.statusSynced {
		b.retestAndMerge()
	}
}

func (b *blocker) retestAndMerge() {
	for _, pool := range b.Pools {
		b.retestAndMergeOnePool(pool)
	}
}

func (b *blocker) retestAndMergeOnePool(pool *PRPool) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	ic := &cicdv1.IntegrationConfig{}
	if err := b.client.Get(context.Background(), pool.NamespacedName, ic); err != nil {
		b.log.WithName("merger").Error(err, "")
		return
	}
	if ic.Spec.MergeConfig == nil {
		return
	}
	log := b.log.WithName("merger").WithValues("repo", genPoolKey(ic))

	gitCli, err := utils.GetGitCli(ic, b.client)
	if err != nil {
		log.Error(err, "")
		return
	}

	// Exit if we're waiting for batch re-test
	if pool.CurrentBatch != nil {
		if err := b.handleBatch(pool, ic, gitCli); err != nil {
			log.Error(err, "")
		}
		return
	}

	// Exit if there is no successful PRs
	if len(pool.MergePool[git.CommitStatusStateSuccess]) == 0 {
		return
	}

	// Sort PRs - older (low id) one is prioritized
	candidates := sortPullRequestByID(pool.MergePool[git.CommitStatusStateSuccess])

	// PR with highest priority (oldest one)
	pr := candidates[0]
	branch := cicdv1.GitRef(pr.Base.Ref).GetBranch()

	// Check commit statuses' base branch
	isBaseLatest, err := checkBaseSHA(branch, ic, pr, gitCli)
	if err != nil {
		log.Error(err, "")
		return
	}

	// Merge it if the tests are done based on the latest commit
	if isBaseLatest {
		log.Info(fmt.Sprintf("Merging PR #%d into %s", pr.ID, branch))
		// TODO - merge commit template
		if err := gitCli.MergePullRequest(pr.ID, pr.Head.Sha, ic.Spec.MergeConfig.Method, ""); err != nil {
			log.Error(err, "")
			return
		}
	} else {
		// If not, retest it!
		log.Info(fmt.Sprintf("PR #%d is not tested based on the latest commit of %s. Retesting", pr.ID, branch))
		pool.CurrentBatch = &Batch{}

		// Collect batches, with same base branch
		for _, p := range candidates {
			if cicdv1.GitRef(p.Base.Ref).GetBranch() != branch {
				break
			}
			pool.CurrentBatch.PRs = append(pool.CurrentBatch.PRs, pr)
			if len(pool.CurrentBatch.PRs) == maxBatchSize {
				break
			}
		}

		// Retest it (create IJ) TODO - batched IJ - not implemented yet
		ij := dispatcher.GeneratePreSubmit(&pr.PullRequest, &git.Repository{Name: ic.Spec.Git.Repository, URL: pr.URL}, &pr.Sender, ic)
		pool.CurrentBatch.Job = types.NamespacedName{Name: ij.Name, Namespace: ij.Namespace}
		if err := b.client.Create(context.Background(), ij); err != nil {
			log.Error(err, "")
			return
		}
	}
}

func (b *blocker) handleBatch(pool *PRPool, ic *cicdv1.IntegrationConfig, gitCli git.Client) error {
	// Check if the tests for batch is successful
	ij := &cicdv1.IntegrationJob{}
	if err := b.client.Get(context.Background(), pool.CurrentBatch.Job, ij); err != nil {
		return err
	}

	switch ij.Status.State {
	case cicdv1.IntegrationJobStateCompleted:
		// If batch test is successful, merge them all, sequentially
		// TODO - what if the target branch is updated during the test...? (manually by a user)
		for _, pr := range pool.CurrentBatch.PRs {
			b.log.WithName("merger").WithValues("integrationconfig", ic.Namespace+"/"+ic.Name).Info(fmt.Sprintf("Merging PR #%d into %s", pr.ID, cicdv1.GitRef(pr.Base.Ref).GetBranch()))
			// TODO - merge commit template
			if err := gitCli.MergePullRequest(pr.ID, pr.Head.Sha, ic.Spec.MergeConfig.Method, ""); err != nil {
				return err
			}
		}
		pool.CurrentBatch = nil
	case cicdv1.IntegrationJobStateFailed:
		// If batch test fails, test again with one less PR in the batch
		// But if the length is 1 and fails...? Kick it out from the merge pool
		if pool.CurrentBatch.Len() <= 1 {
			pool.CurrentBatch = nil
		} else {
			pool.CurrentBatch.PRs = pool.CurrentBatch.PRs[:len(pool.CurrentBatch.PRs)-1]
			pool.CurrentBatch.Job.Name = ""
			pool.CurrentBatch.Job.Namespace = ""
			// TODO - re-test (create batched IJ - not implemented yet)
			// pool.CurrentBatch.Job = types.NamespacedName{Name: ij.Name, Namespace: ij.Namespace}
		}
	default:
		// Do nothing if it's still running
	}
	return nil
}

func checkBaseSHA(baseBranch string, ic *cicdv1.IntegrationConfig, pr *PullRequest, gitCli git.Client) (bool, error) {
	// Base's latest SHA
	branch, err := gitCli.GetBranch(baseBranch)
	if err != nil {
		return false, err
	}
	latest := branch.CommitID

	jobs := dispatcher.FilterJobs(ic.Spec.Jobs.PreSubmit, git.EventTypePullRequest, pr.Base.Ref)
	for _, j := range jobs {
		status, exist := pr.Statuses[j.Name]
		// The status will be there... but if not, it should've been filtered from sync_status
		if !exist {
			continue
		}

		if pipelinemanager.ParseBaseFromDescription(status.Description) != latest {
			return false, nil
		}
	}
	return true, nil
}

func sortPullRequestByID(prs map[int]*PullRequest) PullRequestByID {
	var candidates PullRequestByID
	for _, pr := range prs {
		candidates = append(candidates, pr)
	}
	sort.Sort(candidates)
	return candidates
}

// PullRequestByID is a PR id, sorted by ID
type PullRequestByID []*PullRequest

func (p PullRequestByID) Len() int {
	return len(p)
}

func (p PullRequestByID) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p PullRequestByID) Less(i, j int) bool {
	return p[i].ID < p[j].ID
}
