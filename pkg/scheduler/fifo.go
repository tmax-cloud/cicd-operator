package scheduler

import (
	"context"
	"fmt"
	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/pkg/pipelinemanager"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sort"
)

// fifo is a FIFO scheduler
func (s Scheduler) fifo() {
	// List all IntegrationJobs
	jobList := &cicdv1.IntegrationJobList{}
	if err := s.k8sClient.List(context.TODO(), jobList); err != nil {
		log.Error(err, "")
		return
	}

	// Get running jobs first
	runningJobs := filter(jobList.Items, cicdv1.IntegrationJobStateRunning)

	availableCnt := configs.MaxPipelineRun - len(runningJobs)

	// Check if running jobs are actually running (has pipelineRun, pipelineRun is running)
	for _, j := range runningJobs {
		pr := &tektonv1beta1.PipelineRun{}
		err := s.k8sClient.Get(context.TODO(), types.NamespacedName{Name: pipelinemanager.Name(&j), Namespace: j.Namespace}, pr)
		// If PipelineRun is not found or is already completed, is not actually running
		if (err != nil && errors.IsNotFound(err)) || (err == nil && pr.Status.CompletionTime != nil) {
			availableCnt++
		}
	}

	// If the number of running jobs is greater or equals to the max pipeline run, no scheduling is allowed
	if availableCnt <= 0 {
		log.Info("Max number of PipelineRuns already exist")
		return
	}

	// Now schedule pending jobs
	pendingJobs := filter(jobList.Items, cicdv1.IntegrationJobStatePending)

	// Sort it as FIFO
	sort.Sort(jobListFifo(pendingJobs))

	for _, j := range pendingJobs {
		// Break when max number of running IntegrationJobs exist
		if availableCnt <= 0 {
			break
		}

		// Generate and create PipelineRun
		pr, err := pipelinemanager.Generate(&j)
		if err != nil {
			// TODO - update IntegrationJob status - reason: cannot generate PipelineRun
			log.Error(err, "")
			continue
		}
		if err := controllerutil.SetControllerReference(&j, pr, s.scheme); err != nil {
			// TODO - update IntegrationJob status - reason: cannot create PipelineRun
			log.Error(err, "")
			continue
		}

		// Check if PipelineRun already exists
		if err := s.k8sClient.Get(context.TODO(), types.NamespacedName{Name: pr.Name, Namespace: pr.Namespace}, pr); err != nil {
			// Not found error is expected
			if !errors.IsNotFound(err) {
				// TODO - update IntegrationJob status - reason: cannot get PipelineRun
				log.Error(err, "")
				continue
			}
		} else {
			// PipelineRun already exists...
			availableCnt--
			continue
		}

		log.Info(fmt.Sprintf("Scheduled %s / %s / %s", j.Name, j.Namespace, j.CreationTimestamp))
		// Create PipelineRun only when there is no Pipeline exists
		if err := s.k8sClient.Create(context.TODO(), pr); err != nil {
			// TODO - update IntegrationJob status - reason: cannot create PipelineRun
			log.Error(err, "")
			continue
		}

		availableCnt--
	}
}

// filter jobs with specific state
func filter(from []cicdv1.IntegrationJob, state cicdv1.IntegrationJobState) []cicdv1.IntegrationJob {
	var to []cicdv1.IntegrationJob
	for _, j := range from {
		if j.Status.State == state {
			to = append(to, j)
		}
	}
	return to
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
	xTime := j[x].CreationTimestamp.Time
	yTime := j[y].CreationTimestamp.Time
	// Secondary order is alphanumerical
	if xTime.Equal(yTime) {
		return j[x].Namespace+"-"+j[x].Name < j[y].Namespace+"-"+j[y].Name
	} else {
		// Primary order is time-based
		return xTime.Before(yTime)
	}
}
