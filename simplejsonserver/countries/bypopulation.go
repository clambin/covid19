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
	response, err := getStatsByCountry(handler.CovidDB, req.Args, handler.Mode)
	if err != nil {
		return nil, err
	}

	response = handler.pivotResponse(response)

	var population map[string]int64
	if population, err = handler.PopDB.List(); err != nil {
		return nil, err
	}

	// calculate figure by country population
	for index := range response.Columns[0].Data.(query.TimeColumn) {
		name := response.Columns[1].Data.(query.StringColumn)[index]
		count := response.Columns[2].Data.(query.NumberColumn)[index]

		code, codeFound := fetcher.CountryCodes[name]
		if codeFound == false {
			code = name
		}
		response.Columns[1].Data.(query.StringColumn)[index] = code

		var rate float64
		if pop, popFound := population[code]; popFound {
			rate = count / float64(pop)
		}
		response.Columns[2].Data.(query.NumberColumn)[index] = rate
	}

	// fix column name
	switch handler.Mode {
	case CountryConfirmed:
		response.Columns[2].Text = "confirmed"
	case CountryDeaths:
		response.Columns[2].Text = "deaths"
	}

	return response, nil
}

func (handler *ByCountryByPopulationHandler) pivotResponse(input *query.TableResponse) (output *query.TableResponse) {
	// pivot from:
	// Columns {
	//		timestamp column
	//      data column (text: country name)
	// }
	// to:
	// Columns {
	// 		timestamp column
	//		country code column
	//		data column
	// }

	var (
		timestamps   []time.Time
		countryNames []string
		values       []float64
	)

	timestamp := input.Columns[0].Data.(query.TimeColumn)[0]

	for _, col := range input.Columns[1:] {
		countryName := col.Text
		value := col.Data.(query.NumberColumn)[0]

		timestamps = append(timestamps, timestamp)
		countryNames = append(countryNames, countryName)
		values = append(values, value)
	}

	return &query.TableResponse{
		Columns: []query.Column{
			{Text: "timestamp", Data: query.TimeColumn(timestamps)},
			{Text: "country", Data: query.StringColumn(countryNames)},
			{Text: "???", Data: query.NumberColumn(values)},
		},
	}
}
