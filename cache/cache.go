package cache

import (
	"github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v3/dataset"
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
	totals    *dataset.Dataset
	deltas    *dataset.Dataset
}

// GetTotals return total probe figures up to endTime
func (cache *Cache) GetTotals(endTime time.Time) (totals *dataset.Dataset, err error) {
	err = cache.updateMaybe()
	if err != nil {
		log.WithError(err).Error("failed to retrieve COVID19 entries from the database")
		return
	}

	cache.lock.Lock()
	defer cache.lock.Unlock()
	totals = cache.totals.Copy()
	totals.FilterByRange(time.Time{}, endTime)

	return
}

// GetDeltas returns the incremental probe figures up to endTime
func (cache *Cache) GetDeltas(endTime time.Time) (deltas *dataset.Dataset, err error) {
	err = cache.updateMaybe()
	if err != nil {
		log.WithError(err).Error("failed to retrieve COVID19 entries from the database")
		return
	}

	cache.lock.Lock()
	defer cache.lock.Unlock()

	deltas = cache.deltas.Copy()
	deltas.FilterByRange(time.Time{}, endTime)
	return
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

func GetTotalCases(entries []models.CountryEntry) (result *dataset.Dataset) {
	result = dataset.New()
	for _, entry := range entries {
		result.Add(entry.Timestamp, "confirmed", float64(entry.Confirmed))
		result.Add(entry.Timestamp, "deaths", float64(entry.Deaths))
		//result.Add(entry.Timestamp, "recovered", float64(entry.Recovered))
	}
	return
}

func GetTotalDeltas(totals *dataset.Dataset) (result *dataset.Dataset) {
	result = dataset.New()
	timestamps := totals.GetTimestamps()
	for _, column := range totals.GetColumns() {
		var current float64
		values, _ := totals.GetValues(column)
		for index, value := range values {
			result.Add(timestamps[index], column, value-current)
			current = value
		}
	}
	return
}
