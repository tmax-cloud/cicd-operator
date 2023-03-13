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

package scheduler

import (
	"context"
	"fmt"
	"time"

	tektonv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/pkg/pipelinemanager"
	"github.com/tmax-cloud/cicd-operator/pkg/scheduler/pool"
	"github.com/tmax-cloud/cicd-operator/pkg/structs"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// Scheduler schedules PipelineRun for IntegrationJob

var log = logf.Log.WithName("job-scheduler")

// New is a constructor for a scheduler
func New(c client.Client, s *runtime.Scheme, pm pipelinemanager.PipelineManager) *scheduler {
	log.Info("New scheduler")
	sch := &scheduler{
		k8sClient: c,
		scheme:    s,
		caller:    make(chan struct{}, 1),
		pm:        pm,
	}
	sch.jobPool = pool.New(sch.caller, fifoCompare)
	go sch.start()
	return sch
}

// Scheduler is an interface of scheduler
type Scheduler interface {
	Notify(job *cicdv1.IntegrationJob)
}

// scheduler watches IntegrationJobs and creates corresponding PipelineRuns, considering how many pipeline runs are
// running (in a jobPool)
type scheduler struct {
	k8sClient client.Client
	scheme    *runtime.Scheme

	pm pipelinemanager.PipelineManager

	jobPool pool.JobPool

	// Buffered channel with capacity 1
	// Since scheduler lists resources by itself, the actual scheduling logic should be executed only once even when
	// Schedule is called for several times
	caller chan struct{}
}

// Notify notifies scheduler to sync
func (s scheduler) Notify(job *cicdv1.IntegrationJob) {
	s.jobPool.Lock()
	s.jobPool.SyncJob(job)
	s.jobPool.Unlock()
}

func (s scheduler) start() {
	for range s.caller {
		s.run()
		// Set minimum time gap between scheduling logic
		time.Sleep(3 * time.Second)
	}
}

func (s scheduler) run() {
	s.jobPool.Lock()
	defer s.jobPool.Unlock()
	log.Info("scheduling...")
	availableCnt := configs.MaxPipelineRun - s.jobPool.Running().Len()

	// Check if running jobs are actually running (has pipelineRun, pipelineRun is running)
	s.jobPool.Running().ForEach(s.filterOutRunning(&availableCnt))

	// Check if pending jobs are timeouted
	s.jobPool.Pending().ForEach(s.filterOutPending())

	// If the number of running jobs is greater or equals to the max pipeline run, no scheduling is allowed
	if availableCnt <= 0 {
		log.Info("Max number of PipelineRuns already exist")
		return
	}

	// Schedule if available
	s.jobPool.Pending().ForEach(s.schedulePending(&availableCnt))
}

func (s *scheduler) filterOutRunning(availableCnt *int) func(structs.Item) {
	return func(item structs.Item) {
		j, ok := item.(*pool.JobNode)
		if !ok {
			return
		}
		pr := &tektonv1beta1.PipelineRun{}
		err := s.k8sClient.Get(context.Background(), types.NamespacedName{Name: pipelinemanager.Name(j.IntegrationJob), Namespace: j.Namespace}, pr)
		// If PipelineRun is not found or is already completed, is not actually running
		if (err != nil && errors.IsNotFound(err)) || (err == nil && pr.Status.CompletionTime != nil) {
			*availableCnt = *availableCnt + 1
		}
	}
}

func (s *scheduler) filterOutPending() func(structs.Item) {
	return func(item structs.Item) {
		j, ok := item.(*pool.JobNode)
		if !ok {
			return
		}
		now := time.Now()
		if j.CreationTimestamp.Time.Add(j.Spec.Timeout.Duration).Before(now) {
			msg := fmt.Errorf("integration job %s_%s is failed due to timeout", j.Namespace, j.Name)
			if err := s.patchJobScheduleFailed(j.IntegrationJob, msg.Error()); err != nil {
				log.Error(err, "")
			}
		}
	}
}

func (s *scheduler) schedulePending(availableCnt *int) func(structs.Item) {
	return func(item structs.Item) {
		if *availableCnt <= 0 {
			return
		}
		jobNode, ok := item.(*pool.JobNode)
		if !ok {
			return
		}

		// Check if PipelineRun already exists
		testPr := &tektonv1beta1.PipelineRun{}
		if err := s.k8sClient.Get(context.Background(), types.NamespacedName{Name: pipelinemanager.Name(jobNode.IntegrationJob), Namespace: jobNode.Namespace}, testPr); err != nil {
			// Not found error is expected
			if !errors.IsNotFound(err) {
				log.Error(err, "")
				return
			}
		} else {
			// PipelineRun already exists...
			*availableCnt = *availableCnt - 1
			return
		}

		// Generate PipeLine and PipeLineRun
		pl, pr, err := s.pm.Generate(jobNode.IntegrationJob)

		// Check whether PipeLine exists
		testPl := &tektonv1beta1.Pipeline{}
		if err := s.k8sClient.Get(context.Background(), types.NamespacedName{Name: pipelinemanager.Name(jobNode.IntegrationJob), Namespace: jobNode.Namespace}, testPl); err != nil {
			// If not, create PipeLine
			if err := s.k8sClient.Create(context.Background(), pl); err != nil {
				if err := s.patchJobScheduleFailed(jobNode.IntegrationJob, err.Error()); err != nil {
					log.Error(err, "")
				}
				log.Error(err, "")
				return
			}
		}

		if err != nil {
			if err := s.patchJobScheduleFailed(jobNode.IntegrationJob, err.Error()); err != nil {
				log.Error(err, "")
			}
			log.Error(err, "")
			return
		}
		if err := controllerutil.SetControllerReference(jobNode.IntegrationJob, pr, s.scheme); err != nil {
			if err := s.patchJobScheduleFailed(jobNode.IntegrationJob, err.Error()); err != nil {
				log.Error(err, "")
			}
			log.Error(err, "")
			return
		}

		if err := controllerutil.SetControllerReference(jobNode.IntegrationJob, pl, s.scheme); err != nil {
			if err := s.patchJobScheduleFailed(jobNode.IntegrationJob, err.Error()); err != nil {
				log.Error(err, "")
			}
			log.Error(err, "")
			return
		}

		log.Info(fmt.Sprintf("Scheduled %s / %s / %s", jobNode.Name, jobNode.Namespace, jobNode.CreationTimestamp))
		// Create PipelineRun only when there is no Pipeline exists
		if err := s.k8sClient.Create(context.Background(), pr); err != nil {
			if err := s.patchJobScheduleFailed(jobNode.IntegrationJob, err.Error()); err != nil {
				log.Error(err, "")
			}
			log.Error(err, "")
			return
		}

		*availableCnt = *availableCnt - 1
	}
}

func (s *scheduler) patchJobScheduleFailed(job *cicdv1.IntegrationJob, msg string) error {
	original := job.DeepCopy()

	job.Status.State = cicdv1.IntegrationJobStateFailed
	job.Status.Message = msg
	job.Status.CompletionTime = &metav1.Time{Time: time.Now()}

	p := client.MergeFrom(original)
	return s.k8sClient.Status().Patch(context.Background(), job, p)
}
