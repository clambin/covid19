package scheduler

import (
	"time"
)

// Runnable interface for anything to be managed by the Scheduler
type Runnable interface {
	Run() error
}

// scheduled contains a scheduled probe & its configuration data
type scheduled struct {
	task     Runnable
	interval time.Duration
	nextRun  time.Time
}

func newScheduled (task Runnable, interval time.Duration)  (*scheduled) {
	return &scheduled{task: task, interval: interval}
}

// run a scheduled task
func (scheduledItem *scheduled) Run() error {
	err := scheduledItem.task.Run()
	scheduledItem.nextRun = time.Now().Add(scheduledItem.interval)
	return err
}

// shouldRun checks if a probe be run
func (scheduledItem *scheduled) shouldRun() (bool, time.Duration) {
	if time.Now().After(scheduledItem.nextRun) {
		return true, scheduledItem.interval
	}
	return false, scheduledItem.nextRun.Sub(time.Now())
}

