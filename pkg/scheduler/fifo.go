package scheduler

import (
	"context"
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/pkg/pipelinemanager"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
)

// fifo is a FIFO scheduler
func (s Scheduler) fifo() {
	log.Info("fifo")
	// Get running jobs first
	runningLabels := map[string]string{cicdv1.JobLabelStatus: string(cicdv1.IntegrationJobStateRunning)}
	runningJobList := &cicdv1.IntegrationJobList{}
	if err := s.k8sClient.List(context.TODO(), runningJobList, client.MatchingLabels(runningLabels)); err != nil {
		log.Error(err, "")
		return
	}
	if len(runningJobList.Items) >= configs.MaxPipelineRun {
		return
	}

	// Now schedule pending jobs
	pendingLabels := map[string]string{cicdv1.JobLabelStatus: string(cicdv1.IntegrationJobStatePending)}
	pendingJobList := &cicdv1.IntegrationJobList{}
	if err := s.k8sClient.List(context.TODO(), pendingJobList, client.MatchingLabels(pendingLabels)); err != nil {
		log.Error(err, "")
		return
	}

	// Sort it as FIFO
	sort.Sort(jobListFifo(pendingJobList.Items))

	availableCnt := configs.MaxPipelineRun - len(runningJobList.Items)
	for _, j := range pendingJobList.Items {
		// Break when max number of running IntegrationJobs exist
		if availableCnt <= 0 {
			break
		}

		// TODO - generate PipelineRun
		pr, err := pipelinemanager.Generate(&j)
		if err != nil {
			// TODO - update IntegrationJob status - reason: cannot generate PipelineRun
			log.Error(err, "")
			continue
		}
		if err := s.k8sClient.Create(context.TODO(), pr); err != nil {
			// TODO - update IntegrationJob status - reason: cannot create PipelineRun
			log.Error(err, "")
			continue
		}
		log.Info(fmt.Sprintf("%s / %s / %s", j.Name, j.Namespace, j.CreationTimestamp))

		availableCnt--
	}
}

// FIFO sorter for IntegrationJob
type jobListFifo []cicdv1.IntegrationJob

func (j jobListFifo) Len() int {
	return len(j)
}

func (j jobListFifo) Swap(x, y int) {
	j[x], j[y] = j[y], j[x]
}

func (j jobListFifo) Less(x, y int) bool {
	return j[x].CreationTimestamp.Time.Before(j[y].CreationTimestamp.Time)
}
