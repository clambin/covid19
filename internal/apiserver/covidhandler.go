package apiserver

import (
	"errors"

	"covid19api/internal/coviddb"
)

// CovidAPIHandler implements by business logic for GrafanaAPIHandler
type CovidAPIHandler struct {
	db coviddb.CovidDB
}

// CreateCovidAPIHandler creates a CovidAPIHandler object
func CreateCovidAPIHandler(db coviddb.CovidDB) (*CovidAPIHandler) {
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

	entries, err := apihandler.db.List(params.To) 

	if err != nil {
		return make([]series, 0), err
	}

	return buildTargets(entries, params.Targets), nil
}

// build the requested targets
func buildTargets(entries []coviddb.CountryEntry, targets []string) ([]series) {
	seriesList := make([]series, 0)
	totalcases  := coviddb.GetTotalCases(entries)

	for _, target := range targets {
		switch target {
		case "confirmed":
			seriesList = append(seriesList, series{Target: target, Datapoints: totalcases[coviddb.CONFIRMED]})
		case "confirmed-delta":
			seriesList = append(seriesList, series{Target: target, Datapoints: coviddb.GetTotalDeltas(totalcases[coviddb.CONFIRMED])})
		case "recovered":
			seriesList = append(seriesList, series{Target: target, Datapoints: totalcases[coviddb.RECOVERED]})
		case "recovered-delta":
			seriesList = append(seriesList, series{Target: target, Datapoints: coviddb.GetTotalDeltas(totalcases[coviddb.RECOVERED])})
		case "death":
			seriesList = append(seriesList, series{Target: target, Datapoints: totalcases[coviddb.DEATHS]})
		case "death-delta":
			seriesList = append(seriesList, series{Target: target, Datapoints: coviddb.GetTotalDeltas(totalcases[coviddb.DEATHS])})
		case "active":
			seriesList = append(seriesList, series{Target: target, Datapoints: totalcases[coviddb.ACTIVE]})
		case "active-delta":
			seriesList = append(seriesList, series{Target: target, Datapoints: coviddb.GetTotalDeltas(totalcases[coviddb.ACTIVE])})
		}
	}

	return seriesList
}
