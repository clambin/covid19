package covidcache

import (
	"context"
	"github.com/clambin/covid19/coviddb"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// Cache caches the total evolution of COVID-19 figures
// Helper function to improve responsiveness of Grafana API server
type Cache struct {
	DB      coviddb.DB
	refresh chan struct{}
	lock    sync.RWMutex
	totals  []CacheEntry
	deltas  []CacheEntry
}

// CacheEntry holds one date's data
type CacheEntry struct {
	Timestamp time.Time
	Confirmed int64
	Recovered int64
	Deaths    int64
	Active    int64
}

// New cache
func New(db coviddb.DB) *Cache {
	return &Cache{
		DB:      db,
		refresh: make(chan struct{}),
	}
}

// Run the cache
func (cache *Cache) Run(ctx context.Context) {
	_ = cache.Update()

	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false
		case <-cache.refresh:
			_ = cache.Update()
		}
	}
}

// Refresh asks the cache to refresh itself from the database
func (cache *Cache) Refresh() {
	cache.refresh <- struct{}{}
}

// Update recalculates the cached data
func (cache *Cache) Update() (err error) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	var entries []coviddb.CountryEntry
	if entries, err = cache.DB.List(); err == nil {
		cache.totals = GetTotalCases(entries)
		cache.deltas = GetTotalDeltas(cache.totals)

	}

	if err != nil {
		log.WithField("err", err).Warning("failed to refresh cache")
	}

	return
}

// GetTotals return total covid figures up to endTime
func (cache *Cache) GetTotals(endTime time.Time) (totals []CacheEntry) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()
	return filterEntries(cache.totals, endTime)
}

// GetDeltas returns the incremental covid figures up to endTime
func (cache *Cache) GetDeltas(endTime time.Time) (deltas []CacheEntry) {
	cache.lock.RLock()
	defer cache.lock.RUnlock()
	return filterEntries(cache.deltas, endTime)
}

func filterEntries(entries []CacheEntry, end time.Time) (result []CacheEntry) {
	for _, entry := range entries {
		if entry.Timestamp.After(end) {
			break
		}
		result = append(result, entry)
	}
	return

}
