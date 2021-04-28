package blocker

import (
	"fmt"
	"github.com/go-logr/logr"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"sync"
	"time"
)

const (
	blockerContext        = "blocker"
	defaultBlockerMessage = "Not mergeable."
)

// blocker blocks PRs to be merged. TODO - Need a cool name
// There are 3 main roles for blocker.
//   1. (Pool Syncer) Sync Pools with github/gitlab's open PullRequests list for each v1.IntegrationConfig.
//   2. (Status Syncer) Check commit statuses/merge conflicts for PRs which meet all the conditions of v1.MergeQuery.
//      (We say the PRs are in 'MergePool')
//   3. (Merger) Merge PRs in the merge pool with successful commit statuses and no merge conflicts.
// These three roles run in their own goroutine, periodically.
type blocker struct {
	client client.Client
	log    logr.Logger

	// Pools contains PR pools for each IntegrationConfigs existing in the cluster.
	// It is kind of a cache of PRs
	Pools map[poolKey]*PRPool

	lastPoolSync time.Time

	// poolSynced is a channel from PoolSyncer to StatusSyncer, which indicates the completion of pool sync
	poolSynced chan struct{}

	// statusSynced is a channel from StatusSyncer to Merger, which indicates the completion of status sync
	statusSynced chan struct{}
}

// New creates a new blocker
func New(c client.Client) *blocker {
	return &blocker{
		client:       c,
		log:          logf.Log.WithName("blocker"),
		lastPoolSync: time.Now(),
		poolSynced:   make(chan struct{}, 1),
		statusSynced: make(chan struct{}, 1),
		Pools:        map[poolKey]*PRPool{},
	}
}

// Start executes three main components of the blocker
func (b *blocker) Start() {
	go b.loopSyncPRs()
	go b.loopSyncMergePoolStatus()
	go b.loopMerge()
}

// poolKey is a key for PR pools.
type poolKey string

// genPoolKey generates a default key for the IntegrationConfig (i.e., <APIUrl>/<Repository>)
func genPoolKey(ic *cicdv1.IntegrationConfig) poolKey {
	host := strings.TrimPrefix(ic.Spec.Git.GetAPIUrl(), "http://")
	host = strings.TrimPrefix(host, "https://")
	return poolKey(fmt.Sprintf("%s/%s", host, ic.Spec.Git.Repository))
}

// PRPool is a PullRequest pool(=cache) of an IntegrationConfig
type PRPool struct {
	lock sync.Mutex

	// NamespacedName stores a name and a namespace of source IntegrationConfig
	types.NamespacedName

	// PullRequests store all open PullRequests for an IntegrationConfig, including those who does not meet conditions.
	PullRequests map[int]*PullRequest

	// MergePool is a cache of PRs which meet all the conditions except for commit status/merge conflict conditions
	MergePool MergePool

	// CurrentBatch is a batch of PRs, waiting for a block-merge.
	// If it's non-nil, maybe merger is retesting the PRs.
	CurrentBatch *Batch
}

// Batch is a batch of PRs, waiting for a block-merge.
type Batch struct {
	// PRs in the batch
	PRs []*PullRequest

	// Job is a IntegrationJob's namespaced name for the batch job
	Job types.NamespacedName
}

// Contains checks if a PR is in the batch
func (b *Batch) Contains(id int) bool {
	for _, pr := range b.PRs {
		if pr.ID == id {
			return true
		}
	}
	return false
}

// Len is a length of the batch
func (b *Batch) Len() int {
	return len(b.PRs)
}

// NewPRPool creates a new PRPool
func NewPRPool(ns, name string) *PRPool {
	return &PRPool{
		PullRequests:   map[int]*PullRequest{},
		MergePool:      NewMergePool(),
		NamespacedName: types.NamespacedName{Name: name, Namespace: ns},
	}
}

// MergePool is a pool for PRs.
// Keys are git.CommitStatusState (same as PullRequest.BlockerStatus) - pr.ID
type MergePool map[git.CommitStatusState]map[int]*PullRequest

// NewMergePool creates a new MergePool
func NewMergePool() MergePool {
	prs := map[git.CommitStatusState]map[int]*PullRequest{}
	prs[git.CommitStatusStatePending] = map[int]*PullRequest{}
	prs[git.CommitStatusStateSuccess] = map[int]*PullRequest{}
	return prs
}

// Search searches a given PR from the MergePool.
// Returns nil if it does not exist
func (m MergePool) Search(id int) *PullRequest {
	for _, prs := range m {
		pr, exist := prs[id]
		if exist {
			return pr
		}
	}
	return nil
}

// Add adds a PullRequest to the MergePool
func (m MergePool) Add(pr *PullRequest) {
	m[pr.BlockerStatus][pr.ID] = pr
}

// Delete deletes a given PR from the MergePool
func (m MergePool) Delete(id int) {
	for key := range m {
		_, exist := m[key][id]
		if exist {
			delete(m[key], id)
		}
	}
}

// PullRequest stores a raw git.PullRequest and a cache of the blocker's commit status
type PullRequest struct {
	git.PullRequest

	// BlockerStatus and BlockerDescription is for caching the commit status values
	// Only available statuses are pending and success - no failure/error
	BlockerStatus      git.CommitStatusState
	BlockerDescription string

	// blockerCacheDirty specifies if the commit status should be updated
	blockerCacheDirty bool

	// Statuses stores whole commit statuses of the PR
	Statuses map[string]git.CommitStatus
}
