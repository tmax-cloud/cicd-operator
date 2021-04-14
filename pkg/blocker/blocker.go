package blocker

import (
	"github.com/go-logr/logr"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sync"
)

const (
	blockerContext        = "blocker"
	defaultBlockerMessage = "Not mergeable."
)

// blocker blocks PRs to be merged. TODO - Need a cool name
// There are 3 main roles for blocker.
//   1. Sync Pools with github/gitlab's open PullRequests list for each v1.IntegrationConfig.
//   2. Check commit statuses/merge conflicts for PRs which meet all the conditions of v1.MergeQuery.
//      (We say the PRs are in 'MergePool')
//   3. Merge PRs in the merge pool with successful commit statuses and no merge conflicts.
// These three roles run in their own goroutine, periodically.
type blocker struct {
	client client.Client
	log    logr.Logger

	// Pools contains PR pools for each IntegrationConfigs existing in the cluster.
	// It is kind of a cache of PRs
	Pools map[PoolKey]*PRPool
}

// New creates a new blocker
func New(c client.Client) *blocker {
	return &blocker{
		client: c,
		log:    logf.Log.WithName("blocker"),

		Pools: map[PoolKey]*PRPool{},
	}
}

// Start executes three main components of the blocker
func (b *blocker) Start() {
	go b.loopSyncPRs()
	// TODO - go b.loopSyncMergePoolStatus()
	// TODO - go b.loopMerge()
}

// PoolKey is a key for PR pools.
type PoolKey string

// poolKey generates a default key for the IntegrationConfig (i.e., <Namespace>/<Name>)
func poolKey(ic *cicdv1.IntegrationConfig) PoolKey {
	return PoolKey(ic.Namespace + "/" + ic.Name)
}

// PRPool is a PullRequest pool(=cache) of an IntegrationConfig
type PRPool struct {
	lock sync.Mutex

	// PullRequests store all open PullRequests for an IntegrationConfig, including those who does not meet conditions.
	PullRequests map[int]*PullRequest

	// MergePool is a cache of PRs which meet all the conditions except for commit status/merge conflict conditions
	MergePool MergePool
}

// NewPRPool creates a new PRPool
func NewPRPool() *PRPool {
	return &PRPool{
		PullRequests: map[int]*PullRequest{},
		MergePool:    NewMergePool(),
	}
}

// CheckStatus is a commit status enum
type CheckStatus int

// CheckStatus enum values
const (
	CheckStatusUnknown = iota
	CheckStatusPending
	CheckStatusSuccess
	CheckStatusFailure
	checkStatusLimit
)

// MergePool is a pool for PRs.
// Keys are checkStatus - pr.ID
type MergePool map[CheckStatus]map[int]*PullRequest

// NewMergePool creates a new MergePool
func NewMergePool() MergePool {
	prs := map[CheckStatus]map[int]*PullRequest{}
	for p := CheckStatus(0); p < checkStatusLimit; p++ {
		prs[p] = map[int]*PullRequest{}
	}
	return prs
}

// Search searches a given PR from the MergePool.
// Returns nil if it does not exist
func (m MergePool) Search(id int) *PullRequest {
	var pr *PullRequest
	for p := CheckStatus(0); p < checkStatusLimit; p++ {
		tmp, exist := m[p][id]
		if exist {
			pr = tmp
			break
		}
	}
	return pr
}

// Add adds a PullRequest to the MergePool
func (m MergePool) Add(pr *PullRequest) {
	m[pr.CheckStatus][pr.ID] = pr
}

// Delete deletes a given PR from the MergePool
func (m MergePool) Delete(id int) {
	for p := CheckStatus(0); p < checkStatusLimit; p++ {
		_, exist := m[p][id]
		if exist {
			delete(m[p], id)
		}
	}
}

// PullRequest stores a raw git.PullRequest and a cache of the blocker's commit status
type PullRequest struct {
	git.PullRequest

	BlockerStatus      cicdv1.CommitStatusState
	BlockerDescription string

	CheckStatus
}
