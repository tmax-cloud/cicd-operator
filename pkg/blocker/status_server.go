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
