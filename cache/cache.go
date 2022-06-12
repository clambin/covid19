package cache

import (
	"github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v3/common"
	"github.com/clambin/simplejson/v3/data"
	"github.com/clambin/simplejson/v3/query"
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
	totals    *data.Table
	deltas    *data.Table
}

// GetTotals return total probe figures up to endTime
func (cache *Cache) GetTotals(endTime time.Time) (totals *data.Table, err error) {
	err = cache.updateMaybe()
	if err != nil {
		return
	}

	cache.lock.Lock()
	defer cache.lock.Unlock()
	return cache.totals.Filter(query.Args{Args: common.Args{Range: common.Range{To: endTime}}}), nil
}

// GetDeltas returns the incremental probe figures up to endTime
func (cache *Cache) GetDeltas(endTime time.Time) (deltas *data.Table, err error) {
	err = cache.updateMaybe()
	if err != nil {
		return
	}

	cache.lock.Lock()
	defer cache.lock.Unlock()
	return cache.deltas.Filter(query.Args{Args: common.Args{Range: common.Range{To: endTime}}}), nil
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
		if entries, err = cache.DB.GetTotalsPerDay(); err == nil {
			cache.totals = GetTotalCases(entries)
			cache.deltas = GetTotalDeltas(cache.totals)
		} else {
			cache.once = nil
		}
	})
	return
}

func GetTotalCases(entries []models.CountryEntry) (result *data.Table) {
	timestamps := make([]time.Time, len(entries))
	confirmed := make([]float64, len(entries))
	deaths := make([]float64, len(entries))

	for idx, entry := range entries {
		timestamps[idx] = entry.Timestamp
		confirmed[idx] = float64(entry.Confirmed)
		deaths[idx] = float64(entry.Deaths)
	}
	return data.New(
		data.Column{Name: "timestamp", Values: timestamps},
		data.Column{Name: "confirmed", Values: confirmed},
		data.Column{Name: "deaths", Values: deaths},
	)
}

func GetTotalDeltas(totals *data.Table) (result *data.Table) {
	confirmed, _ := totals.GetFloatValues("confirmed")
	deaths, _ := totals.GetFloatValues("deaths")
	return data.New(
		data.Column{Name: "timestamp", Values: totals.GetTimestamps()},
		data.Column{Name: "confirmed", Values: makeDeltas(confirmed)},
		data.Column{Name: "deaths", Values: makeDeltas(deaths)},
	)
}

func makeDeltas(input []float64) (output []float64) {
	var current float64
	output = make([]float64, len(input))
	for idx, value := range input {
		output[idx] = value - current
		current = value
	}
	return
}
