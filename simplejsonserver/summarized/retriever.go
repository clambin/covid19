package summarized

import (
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v6"
	"github.com/clambin/simplejson/v6/pkg/data"
	"time"
)

type Retriever struct {
	DB db.CovidStore
}

func (r *Retriever) getTotalsForCountry(args simplejson.QueryArgs) (entries []models.CountryEntry, err error) {
	var countryName string
	countryName, err = evaluateAdHocFilter(args.AdHocFilters)

	if err != nil {
		return
	}

	return r.DB.GetAllForCountryName(countryName)
}

func dbEntriesToTable(entries []models.CountryEntry) (table *data.Table) {
	timestamps := make([]time.Time, len(entries))
	confirmed := make([]float64, len(entries))
	deaths := make([]float64, len(entries))

	for idx, entry := range entries {
		timestamps[idx] = entry.Timestamp
		confirmed[idx] = float64(entry.Confirmed)
		deaths[idx] = float64(entry.Deaths)
	}
	return data.New(
		data.Column{Name: "timestamp", Values: timestamps},
		data.Column{Name: "confirmed", Values: confirmed},
		data.Column{Name: "deaths", Values: deaths},
	)
}

func createDeltas(totals *data.Table) (deltas *data.Table) {
	confirmed, _ := totals.GetFloatValues("confirmed")
	deaths, _ := totals.GetFloatValues("deaths")
	return data.New(
		data.Column{Name: "timestamp", Values: totals.GetTimestamps()},
		data.Column{Name: "confirmed", Values: makeDeltas(confirmed)},
		data.Column{Name: "deaths", Values: makeDeltas(deaths)},
	)
}

func makeDeltas(input []float64) (output []float64) {
	var current float64
	output = make([]float64, len(input))
	for idx, value := range input {
		output[idx] = value - current
		current = value
	}
	return
}
