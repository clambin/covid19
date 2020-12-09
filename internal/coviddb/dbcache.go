package coviddb

import (
	"sync"
	"time"
)

// DBCache implements a cache for a Covid DB
type DBCache struct {
	db        DB
	retention time.Duration
	expired   time.Time
	content   []CountryEntry
	lock      sync.Mutex
}

// NewCacheWithDB creates a DB Cache object for a defined DB object
func NewCache(db DB, retention time.Duration) *DBCache {
	return &DBCache{db: db, retention: retention, content: make([]CountryEntry, 0)}
}

// List: get the data from the goroutine
func (dbc *DBCache) List(endTime time.Time) ([]CountryEntry, error) {
	dbc.lock.Lock()
	defer dbc.lock.Unlock()

	// FIXME: if endTime is different, we can't use the cache
	if dbc.expired.After(time.Now()) {
		return dbc.content, nil
	}

	content, err := dbc.db.List(endTime)
	if err == nil {
		dbc.content = content
		dbc.expired = time.Now().Add(dbc.retention)
	}

	return dbc.content, err
}
