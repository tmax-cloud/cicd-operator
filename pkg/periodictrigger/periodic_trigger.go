package periodictrigger

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"github.com/tmax-cloud/cicd-operator/pkg/cron"
	"github.com/tmax-cloud/cicd-operator/pkg/interrupts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defaultTickInterval = time.Minute
)

// PeriodicTrigger schedules periodic jobs.
type PeriodicTrigger struct {
	ID   string
	cron *cron.Cron
	log  logr.Logger
	client.Client
	ic *cicdv1.IntegrationConfig
	context.Context
}

// New is a constructor of a PeriodicTrigger
func New(c client.Client, config *cicdv1.IntegrationConfig, ctx context.Context) *PeriodicTrigger {
	rs := utils.RandomString(5)
	return &PeriodicTrigger{
		ID:      rs,
		cron:    cron.New(),
		log:     logf.Log.WithName("periodic-trigger-" + rs),
		Client:  c,
		ic:      config,
		Context: ctx,
	}
}

// Start starts the cron
func (pt *PeriodicTrigger) Start() error {
	pt.log.Info("Start..")
	//start a cron
	pt.cron.Start()

	tickInterval := defaultTickInterval //TODO : configmap으로 설정할 수 있게 할 것

	interrupts.TickLiteral(func() {
		start := time.Now()
		if err := sync(pt.Client, pt.Context, pt.ic, pt.cron, start); err != nil {
			pt.log.Error(err, "Error syncing periodic jobs.")
		}
		pt.log.Info("Synced periodic jobs", "duration=", time.Since(start))
	}, tickInterval)

	return nil
}

// Stop stops the cron
func (pt *PeriodicTrigger) Stop() {
	pt.log.Info("Stop..")
	pt.cron.Stop()
}

func sync(IntegrationJobClient client.Client, ctx context.Context, ic *cicdv1.IntegrationConfig, cr *cron.Cron, now time.Time) error {
	logger := logf.Log.WithName("periodic_trigger_sync")
	jobs := &cicdv1.IntegrationJobList{}
	if err := IntegrationJobClient.List(ctx, jobs, client.InNamespace(ic.Namespace)); err != nil {
		return fmt.Errorf("error listing Intergrationjobs: %w", err)
	}
	latestJobs := getLatestIntegrationJobsPeriodic(jobs.Items)

	if err := cr.SyncIntegrationConfig(ic); err != nil {
		logger.Error(err, "Error syncing cron jobs.")
	}

	cronTriggers := sets.NewString()
	for _, job := range cr.QueuedJobs() {
		cronTriggers.Insert(job)
	}

	var errs []error
	for _, p := range ic.Spec.Jobs.Periodic {
		j, previousFound := latestJobs[p.Name]
		if p.Cron == "" {
			logger.Info("There is No Cron String")
			return nil
		} else if cronTriggers.Has(p.Name) {
			shouldTrigger := j.IsCompleted()
			if !previousFound || shouldTrigger {
				integrationJob := generatePeriodic(ic, p.Job)
				integrationJob.Namespace = ic.Namespace
				logger.Info("Triggering new run of cron periodic.")
				if err := IntegrationJobClient.Create(ctx, integrationJob); err != nil {
					errs = append(errs, err)
				}
			} else {
				logger.Info("skipping cron periodic")
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to create %d IntegrationJobs: %v", len(errs), errs)
	}

	return nil
}

func generatePeriodic(config *cicdv1.IntegrationConfig, job cicdv1.Job) *cicdv1.IntegrationJob {
	jobID := utils.RandomString(20)
	return &cicdv1.IntegrationJob{
		ObjectMeta: generatePeriodicMeta(config.Name, config.Namespace, job.Name, jobID),
		Spec: cicdv1.IntegrationJobSpec{
			ConfigRef: cicdv1.IntegrationJobConfigRef{
				Name: config.Name,
				Type: cicdv1.JobTypePeriodic,
			},
			ID:         jobID,
			Jobs:       cicdv1.Jobs{job}, // comment (jh) : Periodic은 job별로 Cron을 갖기 때문에, 개별적으로 pipeline을 만들어야 함
			Workspaces: config.Spec.Workspaces,
			Refs: cicdv1.IntegrationJobRefs{
				Repository: config.Spec.Git.Repository,
				Sender: &cicdv1.IntegrationJobSender{ // comment(jh) : Required value임. 우선은 빈 스트링 넣어놓기
					Name:  "",
					Email: "",
				},
			},
			PodTemplate: config.Spec.PodTemplate,
		},
	}
}

func generatePeriodicMeta(cfgName, cfgNamespace, jobName, jobID string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-%s-%s-%s", cfgName, "periodic", jobName, jobID[:5]), // commnet(jh) : cfgName이 너무 길면, 죽고 있음. 다른 네이밍 방안 고민해볼 것
		Namespace: cfgNamespace,
		Labels: map[string]string{
			cicdv1.JobLabelConfig: cfgName,
			cicdv1.JobLabelID:     jobID,
		},
	}
}

// GetLatestIntegrationJobs filters through the provided IntegrationJobs and returns
// a map of jobType jobs to their latest IntegrationJobs.
func getLatestIntegrationJobsPeriodic(ijs []cicdv1.IntegrationJob) map[string]cicdv1.IntegrationJob {
	latestJobs := make(map[string]cicdv1.IntegrationJob)
	for _, ij := range ijs {
		if ij.Spec.ConfigRef.Type != cicdv1.JobTypePeriodic {
			continue
		}

		// Periodic 타입의 ij 같은 경우, Jobs length는 1
		name := ij.Spec.Jobs[0].Container.Name

		_, exist := latestJobs[name]
		if !exist || ij.Status.StartTime.After(latestJobs[name].Status.StartTime.Time) {
			latestJobs[name] = ij
		}
	}
	return latestJobs
}
