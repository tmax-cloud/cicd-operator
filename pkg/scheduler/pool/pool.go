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

package pool

import (
	"fmt"
	"sync"
	"time"

	v1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/structs"
)

// jobPool stores current status of v1.IntegrationJobs, who are in pending status or running status
// All operations for this pool should be done in thread-safe manner, using Lock and Unlock methods
type jobPool struct {
	jobMap jobMap

	pending structs.SortedUniqueList
	running structs.SortedUniqueList

	scheduleChan chan struct{}
	lock         sync.Mutex
}

// JobPool is an interface of jobPool
type JobPool interface {
	Lock()
	Unlock()
	SyncJob(job *v1.IntegrationJob)
	Running() structs.SortedUniqueList
	Pending() structs.SortedUniqueList
}

// New is a constructor for a jobPool
func New(ch chan struct{}, compareFunc structs.CompareFunc) *jobPool {
	return &jobPool{
		jobMap:       jobMap{},
		pending:      structs.NewSortedUniqueQueue(compareFunc),
		running:      structs.NewSortedUniqueQueue(nil),
		scheduleChan: ch,
		lock:         sync.Mutex{},
	}
}

func (j *jobPool) Running() structs.SortedUniqueList {
	return j.running
}

func (j *jobPool) Pending() structs.SortedUniqueList {
	return j.pending
}

// Lock locks jobPool
func (j *jobPool) Lock() {
	j.lock.Lock()
}

// Unlock unlocks jobPool
func (j *jobPool) Unlock() {
	j.lock.Unlock()
}

// SyncJob syncs jobPool with an incoming IntegrationJob job, considering its status
func (j *jobPool) SyncJob(job *v1.IntegrationJob) {
	// If job state is not set, return
	if job.Status.State == "" {
		return
	}

	nodeID := getNodeID(job)

	oldStatus := v1.IntegrationJobState("")
	newStatus := job.Status.State

	// Make / fetch node pointer
	var node *JobNode
	candidate, exist := j.jobMap[nodeID]
	if exist {
		node = candidate
		oldStatus = candidate.Status.State
		candidate.IntegrationJob = job.DeepCopy()
	} else {
		node = &JobNode{
			IntegrationJob: job.DeepCopy(),
		}
	}
	j.jobMap[nodeID] = node

	// If there's deletion timestamp, dismiss it
	if node.DeletionTimestamp != nil {
		j.pending.Delete(node)
		j.running.Delete(node)
		delete(j.jobMap, nodeID)
		j.sendSchedule()
		return
	}

	// If status is not changed, do nothing
	if exist && oldStatus == newStatus {
		return
	}

	// If it is newly created, put it in proper list
	if !exist {
		switch newStatus {
		case v1.IntegrationJobStatePending:
			j.pending.Add(node)
			timeout := job.Spec.Timeout.Duration - time.Since(job.CreationTimestamp.Time)
			go j.manageTimeout(timeout, job)
		case v1.IntegrationJobStateRunning:
			j.running.Add(node)
		}
		j.sendSchedule()
		return
	}

	// Pending -> Running / Failed
	if oldStatus == v1.IntegrationJobStatePending {
		j.pending.Delete(node)
		if newStatus == v1.IntegrationJobStateRunning {
			j.running.Add(node)
		}
		return
	}

	// Running -> The others
	// If it WAS running and not now, dismiss it (it is completed for some reason)
	if oldStatus == v1.IntegrationJobStateRunning {
		j.running.Delete(node)
		if newStatus == v1.IntegrationJobStatePending {
			j.pending.Add(node)
		} else {
			delete(j.jobMap, nodeID)
		}
		j.sendSchedule()
		return
	}
}

func (j *jobPool) manageTimeout(timeout time.Duration, job *v1.IntegrationJob) {
	time.Sleep(timeout)
	j.sendSchedule()
}

func (j *jobPool) sendSchedule() {
	if len(j.scheduleChan) < cap(j.scheduleChan) {
		j.scheduleChan <- struct{}{}
	}
}

// JobNode is a node to be stored in jobMap and jobPool
type JobNode struct {
	*v1.IntegrationJob
}

// Equals implements Item's method
func (f *JobNode) Equals(another structs.Item) bool {
	fj, ok := another.(*JobNode)
	if !ok {
		return false
	}
	if f == nil || fj == nil {
		return false
	}
	return f.Name == fj.Name && f.Namespace == fj.Namespace
}

// DeepCopy implements Item's method
func (f *JobNode) DeepCopy() structs.Item {
	return &JobNode{
		IntegrationJob: f.IntegrationJob.DeepCopy(),
	}
}

func getNodeID(j *v1.IntegrationJob) string {
	return fmt.Sprintf("%s_%s", j.Namespace, j.Name)
}

type jobMap map[string]*JobNode
