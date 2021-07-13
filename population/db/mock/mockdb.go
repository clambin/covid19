package mock

import "sync"

// DB handle
type DB struct {
	data map[string]int64
	lock sync.RWMutex
}

// Create a mockapi population coviddb
func Create(data map[string]int64) *DB {
	return &DB{data: data}
}

// List all data
func (db *DB) List() (map[string]int64, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	newList := make(map[string]int64)
	for key, value := range db.data {
		newList[key] = value
	}

	return newList, nil
}

// Add adds or updates new population figures
func (db *DB) Add(code string, population int64) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.data[code] = population

	return nil
}

func (db *DB) DeleteAll() {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.data = make(map[string]int64)
}
