package mock

import (
	"covid19/internal/covid/apiclient"
)

// API handle
type API struct {
	data map[string]apiclient.CountryStats
}

// New creates a mock population API Client
func New(data map[string]apiclient.CountryStats) *API {
	return &API{data: data}
}

// List all data
func (api *API) GetCountryStats() (map[string]apiclient.CountryStats, error) {
	return api.data, nil
}
