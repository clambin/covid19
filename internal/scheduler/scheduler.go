package scheduler

import (
	"time"
	log     "github.com/sirupsen/logrus"
)

// Scheduler maintains a list of probes to be scheduled at configured intervals
type Scheduler struct {
	probes []*scheduled
}

// NewScheduler creates a scheduler
func NewScheduler () *Scheduler {
	return &Scheduler{probes: make([]*scheduled, 0)}
}

// Register a new probe
func (scheduler *Scheduler) Register(probe Runnable, interval time.Duration) {
	scheduled := newScheduled(probe, interval)
	scheduler.probes = append(scheduler.probes, scheduled)
}

// Run all registered probes
func (scheduler *Scheduler) Run(once bool) {
	for {
		sleepTime, _ := time.ParseDuration("5m")
		for i, probe := range scheduler.probes {
			shouldRun, waitTime := probe.shouldRun()
			log.Debugf("Probe %d: shouldRun: %v, waitTime: %f", i, shouldRun, waitTime.Seconds())
			if shouldRun {
				probe.Run()
			} else if waitTime < sleepTime {
				sleepTime = waitTime
			}
		}
		if once {
			break
		}
		time.Sleep (sleepTime)
	}
}

