package covidcache

import (
	"github.com/clambin/covid19/internal/coviddb"
	log "github.com/sirupsen/logrus"
	"time"
)

// Cache caches the total evolution of covid figures
// Helper function to improve responsiveness of Grafana API server
type Cache struct {
	DB      coviddb.DB
	Refresh chan bool
	Request chan Request

	totals []CacheEntry
	deltas []CacheEntry
}

// Request for latest data
type Request struct {
	Response chan Response
	End      time.Time
	Delta    bool
}

type Response struct {
	Series []CacheEntry
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
		Refresh: make(chan bool),
		Request: make(chan Request),
	}
}

// Run the cache
func (cache *Cache) Run() {
	if err := cache.update(); err != nil {
		log.WithField("err", err).Warning("failed to refresh cache")
	}

	for {
		select {
		case <-cache.Refresh:
			if err := cache.update(); err != nil {
				log.WithField("err", err).Warning("failed to refresh cache")
			}
		case req := <-cache.Request:
			var series []CacheEntry
			switch req.Delta {
			case false:
				series = cache.getTotals(req.End)
			case true:
				series = cache.getDeltas(req.End)
			}
			req.Response <- Response{
				Series: series,
			}
		}
	}
}

// Update recalculates the cached data
func (cache *Cache) update() (err error) {
	var entries []coviddb.CountryEntry

	if entries, err = cache.DB.List(time.Now()); err == nil {
		cache.totals = GetTotalCases(entries)
		cache.deltas = GetTotalDeltas(cache.totals)

	}
	return
}

// GetTotals gets all totals up to the specified date
func (cache *Cache) getTotals(end time.Time) []CacheEntry {
	return filterEntries(cache.totals, end)
}

// GetDeltas gets all deltas up to the specified date
func (cache *Cache) getDeltas(end time.Time) []CacheEntry {
	return filterEntries(cache.deltas, end)
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
