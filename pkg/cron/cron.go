package cron

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-logr/logr"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	cron "gopkg.in/robfig/cron.v2"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// jobStatus is a cache layer for tracking existing cron jobs
type jobStatus struct {
	// entryID is a unique-identifier for each cron entry generated from cronAgent
	entryID cron.EntryID
	// triggered marks if a job has been triggered for the next cron.QueuedJobs() call
	triggered bool
	// cronStr is a cache for job's cron status
	// cron entry will be regenerated if cron string changes from the periodic job
	cronStr string
}

// Cron is a wrapper for cron.Cron
type Cron struct {
	cronAgent *cron.Cron
	jobs      map[string]*jobStatus // key string is job Name
	logger    logr.Logger
	lock      sync.Mutex
}

// New makes a new Cron object
func New() *Cron {
	return &Cron{
		cronAgent: cron.New(),
		jobs:      map[string]*jobStatus{},
		logger:    logf.Log.WithName("cron"),
	}
}

// Start kicks off current cronAgent scheduler
func (c *Cron) Start() {
	c.cronAgent.Start()
}

// Stop pauses current cronAgent scheduler
func (c *Cron) Stop() {
	c.cronAgent.Stop()
}

// QueuedJobs returns a list of jobs that need to be triggered
// and reset triggered in jobStatus
func (c *Cron) QueuedJobs() []string {
	c.lock.Lock()
	defer c.lock.Unlock()

	res := []string{}
	for k, v := range c.jobs {
		if v.triggered {
			res = append(res, k)
		}
		c.jobs[k].triggered = false
	}
	return res
}

// SyncIntegrationConfig syncs current cronAgent with current IntegrationConfig
// which add/delete jobs accordingly.
func (c *Cron) SyncIntegrationConfig(ic *cicdv1.IntegrationConfig) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	periodicNames := sets.NewString()
	for _, p := range ic.Spec.Jobs.Periodic {
		if err := c.addPeriodic(p); err != nil {
			return err
		}
		periodicNames.Insert(p.Name)
	}

	existing := sets.NewString()
	for k := range c.jobs {
		existing.Insert(k)
	}

	var removalErrors []error
	for _, job := range existing.Difference(periodicNames).List() {
		if err := c.removeJob(job); err != nil {
			removalErrors = append(removalErrors, err)
		}
	}

	return utilerrors.NewAggregate(removalErrors)
}

// HasJob returns if a job has been scheduled in cronAgent or not
func (c *Cron) HasJob(name string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, exist := c.jobs[name]
	return exist
}

func (c *Cron) addPeriodic(periodic cicdv1.Periodic) error {
	if periodic.Cron == "" {
		return nil
	}

	if job, exist := c.jobs[periodic.Job.Name]; exist {
		if job.cronStr == periodic.Cron { // 같은 cron으로 등록된 job이 있으면
			return nil
		}
		// job updated, remove old entry
		if err := c.removeJob(periodic.Job.Name); err != nil {
			return err
		}
	}

	if err := c.addJob(periodic.Job.Name, periodic.Cron); err != nil {
		return err
	}

	return nil
}

// addJob adds a cron entry for a job to cronAgent
func (c *Cron) addJob(name, cron string) error {
	id, err := c.cronAgent.AddFunc("TZ=UTC "+cron, func() {
		c.lock.Lock()
		defer c.lock.Unlock()

		c.jobs[name].triggered = true
		c.logger.Info("Triggering cron job", "name", name)
	})

	if err != nil {
		return fmt.Errorf("cronAgent fails to add job %s with cron %s: %w", name, cron, err)
	}

	c.jobs[name] = &jobStatus{
		entryID: id,
		cronStr: cron,
		// try to kick of a periodic trigger right away
		triggered: strings.HasPrefix(cron, "@every"),
	}

	c.logger.Info("Added new cron job", "name", name, "cron", cron)
	return nil
}

// removeJob removes the job from cronAgent
func (c *Cron) removeJob(name string) error {
	job, exist := c.jobs[name]
	if !exist {
		return fmt.Errorf("job %s has not been added to cronAgent yet", name)
	}
	c.cronAgent.Remove(job.entryID)
	delete(c.jobs, name)
	c.logger.Info("Removed previous cron job", "name", name)
	return nil
}
