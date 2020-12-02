package scheduler

import (
	"time"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestScheduler (t *testing.T) {
	var probes []*probe
	interval := []int{1, 5, 9}

	scheduler := NewScheduler()

	for i := 0; i<len(interval); i++ {
		probe := newProbe()
		probes = append(probes, probe)
		scheduler.Register(probe, time.Duration(interval[i]) * time.Second)
	}

	assert.Equal(t, len(interval),    len(scheduler.probes))
	scheduler.Run(true)
	for i:=0; i<len(interval); i++ {
		assert.Equal(t, 1, probes[i].runCount)
	}
	time.Sleep(2 * time.Second)
	scheduler.Run(true)
	assert.Equal(t, 2, probes[0].runCount)
	assert.Equal(t, 1, probes[1].runCount)
	assert.Equal(t, 1, probes[2].runCount)
	time.Sleep(5 * time.Second)
	scheduler.Run(true)
	assert.Equal(t, 3, probes[0].runCount)
	assert.Equal(t, 2, probes[1].runCount)
	assert.Equal(t, 1, probes[2].runCount)
}
