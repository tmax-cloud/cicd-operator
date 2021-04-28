package blocker

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/git"
	"net/http"
	"os"
	"strings"
)

// StatusPort is a default port for Blocker status
const StatusPort = 8808

func (b *blocker) StartBlockerStatusServer() {
	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", StatusPort), b.newRouter()); err != nil {
		b.log.Error(err, "")
		os.Exit(1)
	}
}

func (b *blocker) newRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/status", b.handleStatusList)
	router.PathPrefix("/status").HandlerFunc(b.handleStatus)
	return router
}

func (b *blocker) handleStatusList(w http.ResponseWriter, _ *http.Request) {
	var list []statusListEntity

	for key, pool := range b.Pools {
		list = append(list, statusListEntity{
			Key:               string(key),
			PullRequestLength: len(pool.PullRequests),
			Retesting:         pool.CurrentBatch != nil,
		})
	}

	_ = utils.RespondJSON(w, list)
}

type statusListEntity struct {
	Key               string `json:"key"`
	PullRequestLength int    `json:"pull_request_length"`
	Retesting         bool   `json:"retesting"`
}

func (b *blocker) handleStatus(w http.ResponseWriter, req *http.Request) {
	key := strings.TrimPrefix(req.URL.Path, "/status/")

	pool, exist := b.Pools[poolKey(key)]
	if !exist {
		_ = utils.RespondError(w, http.StatusNotFound, "there is no pr pool for "+key)
		return
	}

	var prs []int
	var poolSuccess []int
	var poolPending []int
	var batch []int

	for id := range pool.PullRequests {
		prs = append(prs, id)
	}
	for id := range pool.MergePool[git.CommitStatusStateSuccess] {
		poolSuccess = append(poolSuccess, id)
	}
	for id := range pool.MergePool[git.CommitStatusStatePending] {
		poolPending = append(poolPending, id)
	}
	if pool.CurrentBatch != nil {
		for _, pr := range pool.CurrentBatch.PRs {
			batch = append(batch, pr.ID)
		}
	}

	_ = utils.RespondJSON(w, statusEntity{
		PullRequests:     prs,
		MergePoolSuccess: poolSuccess,
		MergePoolPending: poolPending,
		Retesting:        pool.CurrentBatch != nil,
		RetestingBatch:   batch,
	})
}

type statusEntity struct {
	PullRequests []int `json:"pull_requests"`

	MergePoolSuccess []int `json:"merge_pool_success"`
	MergePoolPending []int `json:"merge_pool_pending"`

	Retesting      bool  `json:"retesting"`
	RetestingBatch []int `json:"retesting_batch"`
}
