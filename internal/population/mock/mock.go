package mock

// PopulationDB handle
type PopulationDB struct {
	data map[string]int64
}

// Create a mock population db
func Create(data map[string]int64) (*PopulationDB) {
	return &PopulationDB{data: data}
}

// List all data
func (db *PopulationDB) List() (map[string]int64, error) {
	return db.data, nil
}

// Add adds or updates new population figures
func (db *PopulationDB) Add(entries map[string]int64) (error) {
	for key, value := range entries {
		db.data[key] = value
	}

	return nil
}
