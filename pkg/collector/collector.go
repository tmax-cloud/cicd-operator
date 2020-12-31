package collector

import (
	"context"
	"fmt"
	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"gopkg.in/robfig/cron.v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// Collector collects garbage (old IntegrationJobs, PipelineRuns...)

var log = logf.Log.WithName("garbage-collector")

type Collector struct {
	client client.Client

	reconfigureChan chan struct{}

	cron     *cron.Cron
	cronSpec string
	cronId   cron.EntryID
}

func New(c client.Client, ch chan struct{}) (*Collector, error) {
	gc := &Collector{
		client:          c,
		cron:            cron.New(),
		cronSpec:        parseGcPeriod(),
		reconfigureChan: ch,
	}
	id, err := gc.cron.AddFunc(gc.cronSpec, gc.collect)
	if err != nil {
		return nil, err
	}
	gc.cronId = id
	return gc, nil
}

func (c *Collector) Start() {
	log.Info("Starting garbage collector")
	c.cron.Start()

	for range c.reconfigureChan {
		if err := c.reconfigure(); err != nil {
			log.Error(err, "")
		}
	}
}

func (c *Collector) reconfigure() error {
	period := parseGcPeriod()
	if c.cronSpec == period {
		return nil
	}
	c.cron.Stop()
	c.cronSpec = period
	c.cron.Remove(c.cronId)
	id, err := c.cron.AddFunc(c.cronSpec, c.collect)
	if err != nil {
		return err
	}
	c.cronId = id
	c.cron.Entry(c.cronId).Job.Run()

	log.Info(fmt.Sprintf("Garbage collector runs %s", c.cronSpec))
	c.cron.Start()
	return nil
}

func (c *Collector) collect() {
	log.Info("Garbage collector is running...")
	jobList := &cicdv1.IntegrationJobList{}
	if err := c.client.List(context.Background(), jobList); err != nil {
		log.Error(err, "")
		return
	}

	now := time.Now()
	for _, j := range jobList.Items {
		if j.Status.CompletionTime == nil {
			continue
		}

		// Collect if it's ttl is over
		if j.Status.CompletionTime.Time.Add(time.Duration(configs.IntegrationJobTTL) * time.Hour).Before(now) {
			log.Info(fmt.Sprintf("Deleting IntegrationJob %s/%s", j.Namespace, j.Name))
			if err := c.client.Delete(context.Background(), &j); err != nil {
				log.Error(err, "")
				continue
			}
		}
	}
}

func parseGcPeriod() string {
	return fmt.Sprintf("@every %dh", configs.CollectPeriod)
}
