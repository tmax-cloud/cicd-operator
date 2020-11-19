package scheduler

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// Scheduler schedules PipelineRun for IntegrationJob

var log = logf.Log.WithName("job-scheduler")

func New(c client.Client) *Scheduler {
	log.Info("New scheduler")
	sch := &Scheduler{
		k8sClient: c,
		caller:    make(chan int, 1),
	}
	go sch.start()
	return sch
}

type Scheduler struct {
	k8sClient client.Client

	// Buffered channel with capacity 1
	// Since scheduler lists resources by itself, the actual scheduling logic should be executed only once even when
	// Schedule is called for several times
	caller chan int
}

func (s Scheduler) start() {
	for range s.caller {
		s.fifo()
	}
}

// Schedule is an exported function, which is called from the other packages
// It just enqueues a 'schedule job'
func (s Scheduler) Schedule() {
	log.Info("Schedule")
	// Exit if channel buffer is full
	if len(s.caller) == cap(s.caller) {
		return
	}
	s.caller <- 1
}
