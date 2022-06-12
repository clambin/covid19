package countries

import (
	"context"
	"github.com/clambin/covid19/covid/probe/fetcher"
	covidStore "github.com/clambin/covid19/covid/store"
	populationStore "github.com/clambin/covid19/population/store"
	"github.com/clambin/simplejson/v3"
	"github.com/clambin/simplejson/v3/query"
	"time"
)

// ByCountryByPopulationHandler returns the latest stats by country
type ByCountryByPopulationHandler struct {
	CovidDB covidStore.CovidStore
	PopDB   populationStore.PopulationStore
	Mode    int
}

var _ simplejson.Handler = &ByCountryHandler{}

func (handler ByCountryByPopulationHandler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		Query: handler.tableQuery,
	}
}

func (handler *ByCountryByPopulationHandler) tableQuery(_ context.Context, req query.Request) (query.Response, error) {
	d, err := getStatsByCountry(handler.CovidDB, req.Args, handler.Mode)
	if err != nil {
		return nil, err
	}

	var population map[string]int64
	if population, err = handler.PopDB.List(); err != nil {
		return nil, err
	}

	var (
		timestamps []time.Time
		codes      []string
		rates      []float64
	)

	ts := d.GetTimestamps()

	for idx, country := range d.GetColumns() {
		if idx == 0 {
			continue
		}

		code, codeFound := fetcher.CountryCodes[country]
		if !codeFound {
			code = country
		}
		values, found := d.GetFloatValues(country)
		if !found {
			continue
		}

		for index, value := range values {
			var rate float64
			if pop, popFound := population[code]; popFound {
				rate = value / float64(pop)
			}

			timestamps = append(timestamps, ts[index])
			codes = append(codes, code)
			rates = append(rates, rate)
		}
	}

	var title string
	switch handler.Mode {
	case CountryConfirmed:
		title = "confirmed"
	case CountryDeaths:
		title = "deaths"
	}

	return &query.TableResponse{Columns: []query.Column{
		{
			Text: "timestamp",
			Data: query.TimeColumn(timestamps),
		},
		{
			Text: "country",
			Data: query.StringColumn(codes),
		},
		{
			Text: title,
			Data: query.NumberColumn(rates),
		},
	}}, nil
}
