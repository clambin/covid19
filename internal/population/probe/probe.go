package probe

import (
	"covid19/internal/population/db"

	log "github.com/sirupsen/logrus"
)

// Probe handle
type Probe struct {
	APIClient APIClient
	db        db.DB
}

// Create a new Probe handle
func Create(apiKey string, db db.DB) *Probe {
	return &Probe{
		APIClient: NewAPIClient(apiKey),
		db:        db,
	}
}

// Run gets latest data and updates the database
func (probe *Probe) Run() error {
	// Call the API
	population, err := probe.APIClient.GetPopulation()

	if err == nil && len(population) > 0 {
		log.Debugf("Got %d new entries", len(population))

		err = probe.db.Add(population)
	}

	if err != nil {
		log.Warning(err)
	}
	return err
}
