package scheduler

import (
	"time"
	"testing"
	"github.com/stretchr/testify/assert"
)

// probe implements a basic Runnable probe
type probe struct {
	runCount int
}

// newProbe creates a probe 
func newProbe() *probe {
	return &probe{runCount: 0}
}

// Run keeps track of how many times it's been run
func (p *probe) Run() error {
	p.runCount = p.runCount + 1
	return nil
}

func TestScheduled (t *testing.T) {
	probe := newProbe()
	p := newScheduled(probe, time.Duration(2 * time.Second))

	assert.Equal(t, 0,     probe.runCount)
	assert.Equal(t, nil,   p.Run())
	assert.Equal(t, 1,     probe.runCount)
	shouldRun, _ := p.shouldRun()
	assert.Equal(t, false, shouldRun)
	time.Sleep(time.Second * 2)
	shouldRun, _ = p.shouldRun()
	assert.Equal(t, true,  shouldRun)
}
