package probe

import (
	log "github.com/sirupsen/logrus"

	"covid19/internal/population/db"
)

// Probe handle
type Probe struct {
	apiClient APIClient
	db        db.DB
}

// Create a new Probe handle
func Create(apiClient APIClient, db db.DB) *Probe {
	return &Probe{apiClient: apiClient, db: db}
}

// Run gets latest data and updates the database
func (probe *Probe) Run() error {
	// Call the API
	population, err := probe.apiClient.GetPopulation()

	if err == nil && len(population) > 0 {
		log.Debugf("Got %d new entries", len(population))

		err = probe.db.Add(population)
	}

	if err != nil {
		log.Warning(err)
	}
	return err
}
