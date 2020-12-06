package population

import (
	//"time"
	//"net/http"

	log "github.com/sirupsen/logrus"
)

// PopulationProbe handle
type PopulationProbe struct {
	apiClient      *APIClient
	db              PopulationDB
}

// Create a new PopulationProbe handle
func Create(apiClient *APIClient, db PopulationDB) (*PopulationProbe) {
	return &PopulationProbe{apiClient: apiClient, db: db}
}

// Run gets latest data and updates the database
func (probe *PopulationProbe) Run() (error) {
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
