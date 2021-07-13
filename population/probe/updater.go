package probe

import (
	"context"
	log "github.com/sirupsen/logrus"
)

type Updater struct {
	Input   chan interface{}
	Stop    chan struct{}
	Done    chan struct{}
	jobDone chan struct{}
	update  func(ctx context.Context, input interface{})
	maxJobs int
}

type Update struct {
	Code    string
	Country string
}

func NewUpdater(update func(ctx context.Context, input interface{}), maxJobs int) *Updater {
	return &Updater{
		Input:   make(chan interface{}),
		Stop:    make(chan struct{}),
		Done:    make(chan struct{}),
		jobDone: make(chan struct{}, maxJobs),
		update:  update,
		maxJobs: maxJobs,
	}
}

func (updater *Updater) Run(ctx context.Context) {
	var runningJobs int
	var running = true
	var inputs []interface{}

loop:
	for running || runningJobs > 0 {
		select {
		case input := <-updater.Input:
			inputs = append(inputs, input)
		case <-ctx.Done():
			break loop
		case <-updater.Stop:
			running = false
		case <-updater.jobDone:
			runningJobs--
			log.WithField("runningJobs", runningJobs).Debug("job done")
		}

		for len(inputs) > 0 && runningJobs < updater.maxJobs {
			runningJobs++
			log.WithField("runningJobs", runningJobs).Debug("scheduling job")
			go func(input interface{}) {
				updater.update(ctx, input)
				updater.jobDone <- struct{}{}
			}(inputs[0])
			inputs = inputs[1:]
		}
	}

	updater.Done <- struct{}{}

	return
}
