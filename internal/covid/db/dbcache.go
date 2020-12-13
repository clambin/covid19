package db

import (
	"sync"
	"time"
)

// Cache implements a cache for a Covid DB
type Cache struct {
	db            DB
	retention     time.Duration
	expired       time.Time
	cachedEndDate time.Time
	content       []CountryEntry
	lock          sync.Mutex
}

// NewCache creates a Cache object for a defined database object
func NewCache(db DB, retention time.Duration) *Cache {
	return &Cache{db: db, retention: retention, content: make([]CountryEntry, 0)}
}

// List gets the data from the cache if possible, else it gets it from the database
func (dbc *Cache) List(endTime time.Time) ([]CountryEntry, error) {
	dbc.lock.Lock()
	defer dbc.lock.Unlock()

	if dbc.expired.After(time.Now()) && endTime.Equal(dbc.cachedEndDate) {
		return dbc.content, nil
	}

	content, err := dbc.db.List(endTime)
	if err == nil {
		dbc.content = content
		dbc.expired = time.Now().Add(dbc.retention)
		dbc.cachedEndDate = endTime
	}

	return dbc.content, err
}
