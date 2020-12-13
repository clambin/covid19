package mock

// DB handle
type DB struct {
	data map[string]int64
}

// Create a mock population db
func Create(data map[string]int64) *DB {
	return &DB{data: data}
}

// List all data
func (db *DB) List() (map[string]int64, error) {
	return db.data, nil
}

// Add adds or updates new population figures
func (db *DB) Add(entries map[string]int64) error {
	for key, value := range entries {
		db.data[key] = value
	}

	return nil
}
