package scheduler

import (
	"context"
	"fmt"
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
	"time"
)

// Scheduler schedules PipelineRun for IntegrationJob

var log = logf.Log.WithName("job-scheduler")

// New is a constructor for a Scheduler
func New(c client.Client, s *runtime.Scheme, pm *pipelinemanager.PipelineManager) *Scheduler {
	log.Info("New scheduler")
	sch := &Scheduler{
		k8sClient: c,
		scheme:    s,
		caller:    make(chan struct{}, 1),
		pm:        pm,
	}
	sch.jobPool = pool.NewJobPool(sch.caller, fifoCompare)
	go sch.start()
	return sch
}

// Scheduler watches IntegrationJobs and creates corresponding PipelineRuns, considering how many pipeline runs are
// running (in a jobPool)
type Scheduler struct {
	k8sClient client.Client
	scheme    *runtime.Scheme

	pm *pipelinemanager.PipelineManager

	jobPool *pool.JobPool

	// Buffered channel with capacity 1
	// Since scheduler lists resources by itself, the actual scheduling logic should be executed only once even when
	// Schedule is called for several times
	caller chan struct{}
}

// Notify notifies scheduler to sync
func (s Scheduler) Notify(job *cicdv1.IntegrationJob) {
	s.jobPool.Lock()
	s.jobPool.SyncJob(job)
	s.jobPool.Unlock()
}

func (s Scheduler) start() {
	for range s.caller {
		s.run()
		// Set minimum time gap between scheduling logic
		time.Sleep(3 * time.Second)
	}
}

func (s Scheduler) run() {
	s.jobPool.Lock()
	defer s.jobPool.Unlock()
	log.Info("scheduling...")
	availableCnt := configs.MaxPipelineRun - s.jobPool.Running.Len()

	// Check if running jobs are actually running (has pipelineRun, pipelineRun is running)
	s.jobPool.Running.ForEach(s.filterOutRunning(&availableCnt))

	// If the number of running jobs is greater or equals to the max pipeline run, no scheduling is allowed
	if availableCnt <= 0 {
		log.Info("Max number of PipelineRuns already exist")
		return
	}

	// Schedule if available
	s.jobPool.Pending.ForEach(s.schedulePending(&availableCnt))
}

func (s *Scheduler) filterOutRunning(availableCnt *int) func(structs.Item) {
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

func (s *Scheduler) schedulePending(availableCnt *int) func(structs.Item) {
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

		// Generate and create PipelineRun
		pr, err := s.pm.Generate(jobNode.IntegrationJob)
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

func (s *Scheduler) patchJobScheduleFailed(job *cicdv1.IntegrationJob, msg string) error {
	original := job.DeepCopy()

	job.Status.State = cicdv1.IntegrationJobStateFailed
	job.Status.Message = msg
	job.Status.CompletionTime = &metav1.Time{Time: time.Now()}

	p := client.MergeFrom(original)
	return s.k8sClient.Status().Patch(context.Background(), job, p)
}
