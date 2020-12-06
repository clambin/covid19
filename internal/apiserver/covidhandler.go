package apiserver

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"covid19/internal/covid"
)

// CovidAPIHandler implements by business logic for GrafanaAPIHandler
type CovidAPIHandler struct {
	db covid.DB
}

// CreateCovidAPIHandler creates a CovidAPIHandler object
func CreateCovidAPIHandler(db covid.DB) (*CovidAPIHandler) {
	return &CovidAPIHandler{db: db}
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

func (apihandler *CovidAPIHandler) search() ([]string) {
	return targets
}

type series struct {
	Target string           `json:"target"`
	Datapoints [][]int64    `json:"datapoints"`
}

// query the DB and return the requested targets
func (apihandler *CovidAPIHandler) query(params RequestParameters) ([]series, error) {
	if apihandler.db == nil {
		return make([]series, 0), errors.New("no database configured")
	}

	entries, err := apihandler.db.List(params.Range.To)

	if err != nil {
		return make([]series, 0), err
	}

	return buildTargets(entries, params.Targets), nil
}

// build the requested targets
func buildTargets(entries []covid.CountryEntry, targets []struct {Target string}) ([]series) {
	seriesList := make([]series, 0)
	totalcases  := covid.GetTotalCases(entries)

	for _, target := range targets {
		switch target.Target {
		case "confirmed":
			seriesList = append(seriesList, series{Target: target.Target, Datapoints: totalcases[covid.CONFIRMED]})
		case "confirmed-delta":
			seriesList = append(seriesList, series{Target: target.Target, Datapoints: covid.GetTotalDeltas(totalcases[covid.CONFIRMED])})
		case "recovered":
			seriesList = append(seriesList, series{Target: target.Target, Datapoints: totalcases[covid.RECOVERED]})
		case "recovered-delta":
			seriesList = append(seriesList, series{Target: target.Target, Datapoints: covid.GetTotalDeltas(totalcases[covid.RECOVERED])})
		case "death":
			seriesList = append(seriesList, series{Target: target.Target, Datapoints: totalcases[covid.DEATHS]})
		case "death-delta":
			seriesList = append(seriesList, series{Target: target.Target, Datapoints: covid.GetTotalDeltas(totalcases[covid.DEATHS])})
		case "active":
			seriesList = append(seriesList, series{Target: target.Target, Datapoints: totalcases[covid.ACTIVE]})
		case "active-delta":
			seriesList = append(seriesList, series{Target: target.Target, Datapoints: covid.GetTotalDeltas(totalcases[covid.ACTIVE])})
		default:
			log.Warningf("dropping unsupported target: %s", target.Target)
		}
	}

	return seriesList
}
