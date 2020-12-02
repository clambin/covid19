package apiserver

import (
	"errors"
	log        "github.com/sirupsen/logrus"

	"covid19api/coviddb"
)

// Handler

type CovidAPIHandler struct {
	db coviddb.CovidDB
}

func CreateCovidAPIHandler(db coviddb.CovidDB) (CovidAPIHandler) {
	return CovidAPIHandler{db: db}
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

func (apihandler CovidAPIHandler) search() ([]string) {
	return targets
}

type series struct {
	Target string           `json:"target"`
	Datapoints [][]int64    `json:"datapoints"`
}

func (apihandler CovidAPIHandler) query(params RequestParameters) ([]series, error) {
	if  apihandler.db == nil {
		return make([]series, 0), errors.New("No database configured")
	}
	entries, err := apihandler.db.List(params.To)

	if err != nil {
		return make([]series, 0), err
	}

	log.Debugf("Entries in DB: %v", entries)
	log.Debugf("Found %d entries in DB", len(entries))

	output, err := buildSeries(entries, params.Targets), nil

	log.Debugf("Output: %v", output)

	return output, err
}

func buildSeries(entries []coviddb.CountryEntry, targets []string) ([]series) {
	series_list := make([]series, 0)
	totalcases  := coviddb.GetTotalCases(entries)

	for _, target := range targets {
		switch target {
		case "confirmed":
			series_list = append(series_list, series{Target: target, Datapoints: totalcases[coviddb.CONFIRMED]})
		case "confirmed-delta":
			series_list = append(series_list, series{Target: target, Datapoints: coviddb.GetTotalDeltas(totalcases[coviddb.CONFIRMED])})
		case "recovered":
			series_list = append(series_list, series{Target: target, Datapoints: totalcases[coviddb.RECOVERED]})
		case "recovered-delta":
			series_list = append(series_list, series{Target: target, Datapoints: coviddb.GetTotalDeltas(totalcases[coviddb.RECOVERED])})
		case "deaths":
			series_list = append(series_list, series{Target: target, Datapoints: totalcases[coviddb.DEATHS]})
		case "deaths-delta":
			series_list = append(series_list, series{Target: target, Datapoints: coviddb.GetTotalDeltas(totalcases[coviddb.DEATHS])})
		case "active":
			series_list = append(series_list, series{Target: target, Datapoints: totalcases[coviddb.ACTIVE]})
		case "active-delta":
			series_list = append(series_list, series{Target: target, Datapoints: coviddb.GetTotalDeltas(totalcases[coviddb.ACTIVE])})
		}
	}

	return series_list
}
