package mock

// DB handle
type DB struct {
	data map[string]int64
}

// Create a mockapi population coviddb
func Create(data map[string]int64) *DB {
	return &DB{data: data}
}

// List all data
func (db *DB) List() (map[string]int64, error) {
	return db.data, nil
}

// Add adds or updates new population figures
func (db *DB) Add(code string, population int64) error {
	db.data[code] = population

	return nil
}
