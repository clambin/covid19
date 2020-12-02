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
	probe    Runnable
	interval time.Duration
	nextRun  time.Time
}

func newScheduled (p Runnable, interval time.Duration)  (*scheduled) {
	return &scheduled{probe: p, interval: interval}
}

// run a probe
func (probe *scheduled) Run() error {
	err := probe.probe.Run()
	probe.nextRun = time.Now().Add(probe.interval)
	return err
}

// shouldRun checks if a probe be run
func (probe *scheduled) shouldRun() (bool, time.Duration) {
	if time.Now().After(probe.nextRun) {
		return true, 0
	} else {
		return false, probe.nextRun.Sub(time.Now())
	}
}

