package interrupts

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// only one instance of the manager ever exists
var single *manager

func init() {
	m := sync.Mutex{}
	single = &manager{
		c:  sync.NewCond(&m),
		wg: sync.WaitGroup{},
	}
	go handleInterrupt()
}

type manager struct {
	// only one signal handler should be installed, so we use a cond to
	// broadcast to workers that an interrupt has occurred
	c *sync.Cond
	// we record whether we've broadcast in the past
	seenSignal bool
	// we want to ensure that all registered servers and workers get a
	// change to gracefully shut down
	wg sync.WaitGroup
}

// handleInterrupt turns an interrupt into a broadcast for our condition.
// This must be called _first_ before any work is registered with the
// manager, or there will be a deadlock.
func handleInterrupt() {
	signalsLock.Lock()
	sigChan := signals()
	signalsLock.Unlock()
	s := <-sigChan
	logrus.WithField("signal", s).Info("Received signal.")
	single.c.L.Lock()
	single.seenSignal = true
	single.c.Broadcast()
	single.c.L.Unlock()
}

// test initialization will set the signals channel in another goroutine
// so we need to synchronize that in order to not trigger the race detector
// even though we know that init() calls will be serial and the test init()
// will fire first
var signalsLock = sync.Mutex{}

var signalChannel = make(chan os.Signal, 1)

// signals allows for injection of mock signals in testing
var signals = func() <-chan os.Signal {
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	return signalChannel
}

// wait executes the cancel when an interrupt is seen or if one has already
// been handled
func wait(cancel func()) {
	single.c.L.Lock()
	if !single.seenSignal {
		single.c.Wait()
	}
	single.c.L.Unlock()
	cancel()
}

// Tick will do work on a dynamically determined interval until an
// interrupt is received. This function is not blocking. Callers are
// expected to exit only after WaitForGracefulShutdown returns to
// ensure all workers have had time to shut down.
func Tick(work func(), interval func() time.Duration) {
	before := time.Time{} // we want to do work right away
	sig := make(chan int, 1)
	single.wg.Add(1)
	go func() {
		defer single.wg.Done()
		for {
			nextInterval := interval()
			nextTick := before.Add(nextInterval)
			sleep := time.Until(nextTick)
			logrus.WithFields(logrus.Fields{
				"before":   before,
				"interval": nextInterval,
				"sleep":    sleep,
			}).Debug("Resolved next tick interval.")
			select {
			case <-time.After(sleep):
				before = time.Now()
				work()
			case <-sig:
				logrus.Info("Worker shutting down...")
				return
			}
		}
	}()

	go wait(func() {
		sig <- 1
	})
}

// TickLiteral runs Tick with an unchanging interval.
func TickLiteral(work func(), interval time.Duration) {
	Tick(work, func() time.Duration {
		return interval
	})
}
