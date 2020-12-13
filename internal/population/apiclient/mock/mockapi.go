package mock

// API handle
type API struct {
	data map[string]int64
}

// New creates a mock population API Client
func New(data map[string]int64) *API {
	return &API{data: data}
}

// List all data
func (api *API) GetPopulation() (map[string]int64, error) {
	return api.data, nil
}
