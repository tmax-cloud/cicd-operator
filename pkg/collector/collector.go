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

package collector

import (
	"context"
	"fmt"
	"time"

	cicdv1 "github.com/tmax-cloud/cicd-operator/api/v1"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"gopkg.in/robfig/cron.v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("garbage-collector")

// Collector is an interface of collector
type Collector interface {
	Start()
}

// collector collects garbage (old IntegrationJobs, PipelineRuns...)
type collector struct {
	client client.Client

	cron     *cron.Cron
	cronSpec string
	cronID   cron.EntryID

	runGc chan struct{}
}

// New is a constructor of collector
func New(c client.Client) (*collector, error) {
	gc := &collector{
		client:   c,
		cron:     cron.New(),
		cronSpec: parseGcPeriod(),
		runGc:    make(chan struct{}, 1),
	}
	configs.RegisterControllerConfigUpdateChan(gc.runGc)
	id, err := gc.cron.AddFunc(gc.cronSpec, gc.collect)
	if err != nil {
		return nil, err
	}
	gc.cronID = id
	return gc, nil
}

// Start starts the collector
func (c *collector) Start() {
	log.Info("Starting garbage collector")
	c.cron.Start()

	for range c.runGc {
		if err := c.reconfigure(); err != nil {
			log.Error(err, "")
		}
	}
}

func (c *collector) reconfigure() error {
	period := parseGcPeriod()
	if c.cronSpec == period {
		return nil
	}
	c.cron.Stop()
	c.cronSpec = period
	c.cron.Remove(c.cronID)
	id, err := c.cron.AddFunc(c.cronSpec, c.collect)
	if err != nil {
		return err
	}
	c.cronID = id
	c.cron.Entry(c.cronID).Job.Run()

	log.Info(fmt.Sprintf("Garbage collector runs %s", c.cronSpec))
	c.cron.Start()
	return nil
}

func (c *collector) collect() {
	log.Info("Garbage collector is running...")
	jobList := &cicdv1.IntegrationJobList{}
	if err := c.client.List(context.Background(), jobList); err != nil {
		if _, ok := err.(*cache.ErrCacheNotStarted); !ok {
			log.Error(err, "")
		}
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
