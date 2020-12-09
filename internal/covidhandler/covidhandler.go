package covidhandler

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	"covid19/internal/covid"
	"covid19/internal/coviddb"
	"covid19/pkg/grafana/apiserver"
)

// APIHandler implements business logic for APIServer
type APIHandler struct {
	dbc *coviddb.DBCache
}

// Create a CovidAPIHandler object
func Create(db coviddb.DB) (*APIHandler, error) {
	if db == nil {
		return nil, errors.New("no database specified")
	}
	return &APIHandler{dbc: coviddb.NewCache(db, 60*time.Second)}, nil
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
	if apiHandler.dbc == nil {
		return make([]apiserver.APIQueryResponse, 0), errors.New("no database configured")
	}

	entries, err := apiHandler.dbc.List(request.Range.To)

	if err != nil {
		return make([]apiserver.APIQueryResponse, 0), err
	}

	return buildTargets(entries, request.Targets), nil
}

// build the requested targets
func buildTargets(entries []coviddb.CountryEntry, targets []struct{ Target string }) []apiserver.APIQueryResponse {
	seriesList := make([]apiserver.APIQueryResponse, 0)
	totalCases := covid.GetTotalCases(entries)

	for _, target := range targets {
		switch target.Target {
		case "confirmed":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: totalCases[covid.CONFIRMED]})
		case "confirmed-delta":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: covid.GetTotalDeltas(totalCases[covid.CONFIRMED])})
		case "recovered":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: totalCases[covid.RECOVERED]})
		case "recovered-delta":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: covid.GetTotalDeltas(totalCases[covid.RECOVERED])})
		case "death":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: totalCases[covid.DEATHS]})
		case "death-delta":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: covid.GetTotalDeltas(totalCases[covid.DEATHS])})
		case "active":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: totalCases[covid.ACTIVE]})
		case "active-delta":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: covid.GetTotalDeltas(totalCases[covid.ACTIVE])})
		default:
			log.Warningf("dropping unsupported target: %s", target.Target)
		}
	}

	return seriesList
}
