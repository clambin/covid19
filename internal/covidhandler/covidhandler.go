package covidhandler

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"covid19/pkg/grafana/apiserver"
	"covid19/internal/covid"
)

// APIHandler implements business logic for APIServer
type APIHandler struct {
	db covid.DB
}

// Create a CovidAPIHandler object
func Create(db covid.DB) (*APIHandler) {
	return &APIHandler{db: db}
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
func (apihandler *APIHandler) Search() ([]string) {
	return targets
}

// Query the DB and return the requested targets
func (apihandler *APIHandler) Query(request *apiserver.APIQueryRequest) ([]apiserver.APIQueryResponse, error) {
	if apihandler.db == nil {
		return make([]apiserver.APIQueryResponse, 0), errors.New("no database configured")
	}

	entries, err := apihandler.db.List(request.Range.To)

	if err != nil {
		return make([]apiserver.APIQueryResponse, 0), err
	}

	return buildTargets(entries, request.Targets), nil
}

// build the requested targets
func buildTargets(entries []covid.CountryEntry, targets []struct {Target string}) ([]apiserver.APIQueryResponse) {
	seriesList := make([]apiserver.APIQueryResponse, 0)
	totalcases  := covid.GetTotalCases(entries)

	for _, target := range targets {
		switch target.Target {
		case "confirmed":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: totalcases[covid.CONFIRMED]})
		case "confirmed-delta":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: covid.GetTotalDeltas(totalcases[covid.CONFIRMED])})
		case "recovered":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: totalcases[covid.RECOVERED]})
		case "recovered-delta":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: covid.GetTotalDeltas(totalcases[covid.RECOVERED])})
		case "death":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: totalcases[covid.DEATHS]})
		case "death-delta":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: covid.GetTotalDeltas(totalcases[covid.DEATHS])})
		case "active":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: totalcases[covid.ACTIVE]})
		case "active-delta":
			seriesList = append(seriesList, apiserver.APIQueryResponse{Target: target.Target, Datapoints: covid.GetTotalDeltas(totalcases[covid.ACTIVE])})
		default:
			log.Warningf("dropping unsupported target: %s", target.Target)
		}
	}

	return seriesList
}
