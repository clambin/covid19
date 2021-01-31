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
	var (
		err        error
		population map[string]int64
	)

	if population, err = probe.APIClient.GetPopulation(); err == nil && len(population) > 0 {
		log.WithField("population", len(population)).Debug("populationProbe got new entries")
		err = probe.db.Add(population)
	}

	return err
}
