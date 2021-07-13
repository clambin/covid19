package probe_test

import (
	"context"
	"github.com/clambin/covid19/population/probe"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type Resource struct {
	counter int
	lock    sync.RWMutex
}

func (resource *Resource) update(_ context.Context, _ interface{}) {
	resource.lock.Lock()
	defer resource.lock.Unlock()
	time.Sleep(10 * time.Millisecond)
	resource.counter++
}

func (resource *Resource) value() int {
	resource.lock.RLock()
	defer resource.lock.RUnlock()
	return resource.counter
}

func TestUpdater_Run(t *testing.T) {
	r := &Resource{}
	updater := probe.NewUpdater(r.update, 5)

	ctx, cancel := context.WithCancel(context.Background())
	go updater.Run(ctx)

	for i := 0; i < 20; i++ {
		updater.Input <- ""
	}
	updater.Stop <- struct{}{}

	<-updater.Done
	assert.Equal(t, 20, r.value())

	cancel()
}

func TestUpdater_Cancel(t *testing.T) {
	r := &Resource{}
	updater := probe.NewUpdater(r.update, 5)

	ctx, cancel := context.WithCancel(context.Background())
	go updater.Run(ctx)

	cancel()
	assert.Eventually(t, func() bool {
		<-updater.Done
		return true
	}, 500*time.Millisecond, 10*time.Millisecond)
}
