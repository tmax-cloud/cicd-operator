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
	"bytes"
	"context"
	"fmt"
	"sort"
	"text/template"
	"time"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/dispatcher"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"github.com/tmax-cloud/cicd-operator/pkg/pipelinemanager"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("blocker")

const (
	maxBatchSize = 10
)

func (b *blocker) loopMerge() {
	// Call retestAndMerge method whenever a status sync is done
	for range b.statusSynced {
		b.retestAndMerge()
	}
}

func (b *blocker) retestAndMerge() {
	for _, pool := range b.Pools {
		go b.retestAndMergeOnePool(pool)
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
		// Do nothing if it's still processing
		if pool.CurrentBatch.Processing {
			log.Info("Handling is in process.. waiting for the completion", "pool", pool.Name)
			return
		}

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

	// PR with the highest priority (the oldest one)
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
		if err := b.mergePullRequest(pr, ic, gitCli); err != nil {
			log.Error(err, "")
			return
		}
	} else {
		// If not, retest it!
		log.Info(fmt.Sprintf("PR #%d is not tested based on the latest commit of %s. Retesting", pr.ID, branch))
		pool.CurrentBatch = &Batch{}

		// Collect batches, with same base branch
		var prIDs []int
		for _, p := range candidates {
			if cicdv1.GitRef(p.Base.Ref).GetBranch() != branch {
				continue
			}
			pool.CurrentBatch.PRs = append(pool.CurrentBatch.PRs, p)
			prIDs = append(prIDs, p.ID)
			if len(pool.CurrentBatch.PRs) == maxBatchSize {
				break
			}
		}

		// Retest it (create IJ)
		log.Info(fmt.Sprintf("Batched tests - %+v", prIDs))
		gitPRs := getGitPRsFromPRs(pool.CurrentBatch.PRs)
		if err := b.createIntegrationJobForBatch(gitPRs, ic, &pool.CurrentBatch.Job); err != nil {
			log.Error(err, "Fail to create integrationJob for batch.")
			return
		}
	}
}

func (b *blocker) handleBatch(pool *PRPool, ic *cicdv1.IntegrationConfig, gitCli git.Client) error {
	pool.CurrentBatch.Processing = true
	defer func() {
		if pool.CurrentBatch != nil {
			pool.CurrentBatch.Processing = false
		}
	}()

	// Check if the tests for batch is successful
	ij := &cicdv1.IntegrationJob{}
	if err := b.client.Get(context.Background(), pool.CurrentBatch.Job, ij); err != nil {
		return err
	}

	switch ij.Status.State {
	case cicdv1.IntegrationJobStateCompleted:
		// If batch test is successful, merge them all, sequentially
		// TODO - what if the target branch is updated during the test...? (manually by a user)
		for len(pool.CurrentBatch.PRs) > 0 {
			if err := b.tryMerge(pool.CurrentBatch.PRs[0], ic, gitCli); err != nil {
				return err
			}
			pool.CurrentBatch.PRs = pool.CurrentBatch.PRs[1:]

			// Wait 5 sec and give github/gitlab time to recalculate the mergeability
			if len(pool.CurrentBatch.PRs) > 0 {
				time.Sleep(5 * time.Second)
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
			gitPRs := getGitPRsFromPRs(pool.CurrentBatch.PRs)
			if err := b.createIntegrationJobForBatch(gitPRs, ic, &pool.CurrentBatch.Job); err != nil {
				log.Error(err, "Fail to create integrationJob for batch.")
				return err
			}
		}
	default:
		// Do nothing if it's still running
	}
	return nil
}

func (b *blocker) tryMerge(pr *PullRequest, ic *cicdv1.IntegrationConfig, gitCli git.Client) error {
	var err error
	const maxRetry = 3
	backOff := 4 * time.Second

	log := b.log.WithName("merger").
		WithValues("repo", genPoolKey(ic)).
		WithValues("ref", pr.Base.Ref).
		WithValues("id", pr.ID)

	for retry := 0; retry < maxRetry; retry++ {
		if retry != 0 {
			log.Info("Retrying...")
		}
		err = b.mergePullRequest(pr, ic, gitCli)
		if err == nil {
			return nil
		}

		log.Info("Error while merging...", "error", err.Error())

		// Sleep & double backOff
		if retry < maxRetry-1 {
			time.Sleep(backOff)
			backOff *= 2
		}
	}

	return err
}

func getGitPRsFromPRs(prs []*PullRequest) []git.PullRequest {
	gitPRs := []git.PullRequest{}
	for _, p := range prs {
		gitPRs = append(gitPRs, p.PullRequest)
	}
	return gitPRs
}

func (b *blocker) createIntegrationJobForBatch(prs []git.PullRequest, ic *cicdv1.IntegrationConfig, batchJob *types.NamespacedName) error {
	// The PRs in batch are assumed to have the same 'repo'.
	dummy := git.User{Name: "tmax-cicd-bot", Email: "bot@cicd.tmax.io"}
	ij := dispatcher.GeneratePreSubmit(prs, &git.Repository{Name: ic.Spec.Git.Repository, URL: prs[0].URL}, &dummy, ic)
	*batchJob = types.NamespacedName{Name: ij.Name, Namespace: ij.Namespace}
	if err := b.client.Create(context.Background(), ij); err != nil {
		log.Error(err, "")
		return err
	}
	return nil
}

func (b *blocker) mergePullRequest(pr *PullRequest, ic *cicdv1.IntegrationConfig, gitCli git.Client) error {
	log := b.log.WithName("merger").WithValues("repo", genPoolKey(ic))
	log.Info(fmt.Sprintf("Merging PR #%d into %s", pr.ID, cicdv1.GitRef(pr.Base.Ref).GetBranch()))

	// Compile commit message
	commitMsg := ""
	if ic.Spec.MergeConfig.CommitTemplate != "" {
		var err error
		// List commits of the pull request
		pr.Commits, err = gitCli.ListPullRequestCommits(pr.ID)
		if err != nil {
			return err
		}

		tmpl := template.New("")
		tmpl, err = tmpl.Parse(ic.Spec.MergeConfig.CommitTemplate)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, pr); err != nil {
			return err
		}
		commitMsg = buf.String()
	}
	if err := gitCli.MergePullRequest(pr.ID, pr.Head.Sha, getMergeMethod(pr, ic), commitMsg); err != nil {
		return err
	}
	return nil
}

func getMergeMethod(pr *PullRequest, ic *cicdv1.IntegrationConfig) git.MergeMethod {
	method := ic.Spec.MergeConfig.Method
	if method == "" {
		method = git.MergeMethodMerge
	}

	// Check squash/merge label
	for _, l := range pr.Labels {
		if configs.MergeKindSquashLabel != "" && l.Name == configs.MergeKindSquashLabel {
			return git.MergeMethodSquash
		}
		if configs.MergeKindMergeLabel != "" && l.Name == configs.MergeKindMergeLabel {
			return git.MergeMethodMerge
		}
	}

	return method
}

func checkBaseSHA(baseBranch string, ic *cicdv1.IntegrationConfig, pr *PullRequest, gitCli git.Client) (bool, error) {
	// Base's latest SHA
	branch, err := gitCli.GetBranch(baseBranch)
	if err != nil {
		return false, err
	}
	latest := branch.CommitID

	jobs := dispatcher.FilterJobs(ic.Spec.Jobs.PreSubmit, git.EventTypePullRequest, pr.Base.Ref, ic.Spec.GlobalBranch)
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
