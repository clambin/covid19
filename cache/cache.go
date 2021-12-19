package cache

import (
	"github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/models"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// Cache caches the total evolution of COVID-19 figures
// Helper function to improve responsiveness of Grafana API server
type Cache struct {
	DB        store.CovidStore
	Retention time.Duration
	lock      sync.Mutex
	expiry    time.Time
	once      *sync.Once
	totals    []Entry
	deltas    []Entry
}

// Entry holds one date's data
type Entry struct {
	Timestamp time.Time
	Confirmed int64
	Recovered int64
	Deaths    int64
	Active    int64
}

// GetTotals return total probe figures up to endTime
func (cache *Cache) GetTotals(endTime time.Time) (totals []Entry, err error) {
	err = cache.updateMaybe()
	if err != nil {
		log.WithError(err).Error("failed to retrieve COVID19 entries from the database")
		return
	}

	cache.lock.Lock()
	defer cache.lock.Unlock()
	return filterEntries(cache.totals, endTime), nil
}

// GetDeltas returns the incremental probe figures up to endTime
func (cache *Cache) GetDeltas(endTime time.Time) (deltas []Entry, err error) {
	err = cache.updateMaybe()
	if err != nil {
		log.WithError(err).Error("failed to retrieve COVID19 entries from the database")
		return
	}

	cache.lock.Lock()
	defer cache.lock.Unlock()
	return filterEntries(cache.deltas, endTime), nil
}

func (cache *Cache) updateMaybe() (err error) {
	cache.lock.Lock()
	defer cache.lock.Unlock()
	if cache.once == nil || time.Now().After(cache.expiry) {
		cache.once = &sync.Once{}
		cache.expiry = time.Now().Add(cache.Retention)
	}
	cache.once.Do(func() {
		var entries []models.CountryEntry
		if entries, err = cache.DB.GetAll(); err == nil {
			cache.totals = GetTotalCases(entries)
			cache.deltas = GetTotalDeltas(cache.totals)
		} else {
			cache.once = nil
		}
	})
	return
}

func filterEntries(entries []Entry, end time.Time) (result []Entry) {
	for _, entry := range entries {
		if entry.Timestamp.After(end) {
			break
		}
		result = append(result, entry)
	}
	return

}
