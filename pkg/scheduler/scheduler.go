package scheduler

import (
	"time"
	log     "github.com/sirupsen/logrus"
)

// Scheduler maintains a list of scheduledItems to be run at configured intervals
type Scheduler struct {
	scheduledItems []*scheduled
}

// NewScheduler creates a scheduler
func NewScheduler () *Scheduler {
	return &Scheduler{scheduledItems: make([]*scheduled, 0)}
}

// Register a new probe
func (scheduler *Scheduler) Register(probe Runnable, interval time.Duration) {
	scheduled := newScheduled(probe, interval)
	scheduler.scheduledItems = append(scheduler.scheduledItems, scheduled)
}

// Run all registered scheduledItems
func (scheduler *Scheduler) Run(once bool) {
	for {
		sleepTime, _ := time.ParseDuration("5m")
		for i, scheduledItem := range scheduler.scheduledItems {
			shouldRun, waitTime := scheduledItem.shouldRun()
			log.Debugf("scheduledItem %d: shouldRun: %v, waitTime: %f", i, shouldRun, waitTime.Seconds())
			if shouldRun {
				scheduledItem.Run()
			}
			if waitTime < sleepTime {
				sleepTime = waitTime
			}
		}
		if once {
			break
		}
		time.Sleep (sleepTime)
	}
}

