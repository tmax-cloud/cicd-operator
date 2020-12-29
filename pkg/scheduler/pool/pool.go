package pool

import (
	"fmt"
	v1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/pkg/structs"
	"sync"
)

// JobPool stores current status of v1.IntegrationJobs, who are in Pending status or Running status
// All operations for this pool should be done in thread-safe manner, using Lock and Unlock methods
type JobPool struct {
	jobMap jobMap

	Pending *structs.SortedUniqueList
	Running *structs.SortedUniqueList

	scheduleChan chan struct{}
	lock         sync.Mutex
}

func NewJobPool(ch chan struct{}, compareFunc structs.CompareFunc) *JobPool {
	return &JobPool{
		jobMap:       jobMap{},
		Pending:      structs.NewSortedUniqueQueue(compareFunc),
		Running:      structs.NewSortedUniqueQueue(nil),
		scheduleChan: ch,
		lock:         sync.Mutex{},
	}
}

func (j *JobPool) Lock() {
	j.lock.Lock()
}

func (j *JobPool) Unlock() {
	j.lock.Unlock()
}

func (j *JobPool) SyncJob(job *v1.IntegrationJob) {
	// If job state is not set, return
	if job.Status.State == "" {
		return
	}

	nodeId := getNodeID(job)

	oldStatus := v1.IntegrationJobState("")
	newStatus := job.Status.State

	// Make / fetch node pointer
	var node *JobNode
	candidate, exist := j.jobMap[nodeId]
	if exist {
		node = candidate
		oldStatus = candidate.Status.State
		candidate.IntegrationJob = job.DeepCopy()
	} else {
		node = &JobNode{
			IntegrationJob: job.DeepCopy(),
		}
	}
	j.jobMap[nodeId] = node

	// If there's deletion timestamp, dismiss it
	if node.DeletionTimestamp != nil {
		j.Pending.Delete(node)
		j.Running.Delete(node)
		delete(j.jobMap, nodeId)
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
			j.Pending.Add(node)
		case v1.IntegrationJobStateRunning:
			j.Running.Add(node)
		}
		j.sendSchedule()
		return
	}

	// Pending -> Running
	if oldStatus == v1.IntegrationJobStatePending {
		j.Pending.Delete(node)
		if newStatus == v1.IntegrationJobStateRunning {
			j.Running.Add(node)
		}
		return
	}

	// Running -> The others
	// If it WAS running and not now, dismiss it (it is completed for some reason)
	if oldStatus == v1.IntegrationJobStateRunning {
		j.Running.Delete(node)
		// TODO : do we need to handle Running -> Pending ??? might not happen...
		if newStatus == v1.IntegrationJobStatePending {
			j.Pending.Add(node)
		} else {
			delete(j.jobMap, nodeId)
		}
		j.sendSchedule()
		return
	}
}

func (j *JobPool) sendSchedule() {
	if len(j.scheduleChan) < cap(j.scheduleChan) {
		j.scheduleChan <- struct{}{}
	}
}

// JobNode is a node to be stored in jobMap and JobPool
type JobNode struct {
	*v1.IntegrationJob
}

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

func (f *JobNode) DeepCopy() structs.Item {
	return &JobNode{
		IntegrationJob: f.IntegrationJob.DeepCopy(),
	}
}

func getNodeID(j *v1.IntegrationJob) string {
	return fmt.Sprintf("%s_%s", j.Namespace, j.Name)
}

type jobMap map[string]*JobNode
