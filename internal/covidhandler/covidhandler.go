package covidhandler

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	"covid19/internal/covid/db"
	"covid19/pkg/grafana/apiserver"
)

// APIHandler implements business logic for APIServer
type APIHandler struct {
	cache *db.Cache
}

// Create a CovidAPIHandler object
func Create(dbh db.DB) (*APIHandler, error) {
	if dbh == nil {
		return nil, errors.New("no database specified")
	}
	return &APIHandler{cache: db.NewCache(dbh, 60*time.Second)}, nil
}

var (
	targets = []string{
		"active",
		"active-delta",
		"confirmed",
		"confirmed-delta",
		"death",
		"death-delta",
		"recovered",
		"recovered-delta",
	}
)

// Search returns all supported targets
func (apiHandler *APIHandler) Search() []string {
	return targets
}

// Query the DB and return the requested targets
func (apiHandler *APIHandler) Query(request *apiserver.APIQueryRequest) ([]apiserver.APIQueryResponse, error) {
	entries, err := apiHandler.cache.List(request.Range.To)

	return buildTargets(entries, request.Targets), err
}

// build the requested targets
func buildTargets(entries []db.CountryEntry, targets []struct{ Target string }) []apiserver.APIQueryResponse {
	seriesList := make([]apiserver.APIQueryResponse, 0)
	totalCases := GetTotalCases(entries)

	for _, target := range targets {
		switch target.Target {
		case "confirmed":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, DataPoints: totalCases[CONFIRMED]})
		case "confirmed-delta":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, DataPoints: GetTotalDeltas(totalCases[CONFIRMED])})
		case "recovered":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, DataPoints: totalCases[RECOVERED]})
		case "recovered-delta":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, DataPoints: GetTotalDeltas(totalCases[RECOVERED])})
		case "death":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, DataPoints: totalCases[DEATHS]})
		case "death-delta":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, DataPoints: GetTotalDeltas(totalCases[DEATHS])})
		case "active":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, DataPoints: totalCases[ACTIVE]})
		case "active-delta":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, DataPoints: GetTotalDeltas(totalCases[ACTIVE])})
		default:
			log.Warningf("dropping unsupported target: %s", target.Target)
		}
	}

	return seriesList
}
