package countries

import (
	"context"
	"github.com/clambin/covid19/covid/probe/fetcher"
	covidStore "github.com/clambin/covid19/covid/store"
	populationStore "github.com/clambin/covid19/population/store"
	"github.com/clambin/simplejson"
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
		TableQuery: handler.tableQuery,
	}
}

func (handler *ByCountryByPopulationHandler) tableQuery(_ context.Context, args *simplejson.TableQueryArgs) (response *simplejson.TableQueryResponse, err error) {
	if response, err = getStatsByCountry(handler.CovidDB, args, handler.Mode); err != nil {
		return
	}

	response = handler.pivotResponse(response)

	var population map[string]int64
	if population, err = handler.PopDB.List(); err != nil {
		return
	}

	// calculate figure by country population
	for index := range response.Columns[0].Data.(simplejson.TableQueryResponseTimeColumn) {
		name := response.Columns[1].Data.(simplejson.TableQueryResponseStringColumn)[index]
		count := response.Columns[2].Data.(simplejson.TableQueryResponseNumberColumn)[index]

		code, codeFound := fetcher.CountryCodes[name]
		if codeFound == false {
			code = name
		}
		response.Columns[1].Data.(simplejson.TableQueryResponseStringColumn)[index] = code

		var rate float64
		if pop, popFound := population[code]; popFound {
			rate = count / float64(pop)
		}
		response.Columns[2].Data.(simplejson.TableQueryResponseNumberColumn)[index] = rate
	}

	// fix column name
	switch handler.Mode {
	case CountryConfirmed:
		response.Columns[2].Text = "confirmed"
	case CountryDeaths:
		response.Columns[2].Text = "deaths"
	}

	return
}

func (handler *ByCountryByPopulationHandler) pivotResponse(input *simplejson.TableQueryResponse) (output *simplejson.TableQueryResponse) {
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

	timestamp := input.Columns[0].Data.(simplejson.TableQueryResponseTimeColumn)[0]

	for _, col := range input.Columns[1:] {
		countryName := col.Text
		value := col.Data.(simplejson.TableQueryResponseNumberColumn)[0]

		timestamps = append(timestamps, timestamp)
		countryNames = append(countryNames, countryName)
		values = append(values, value)
	}

	return &simplejson.TableQueryResponse{
		Columns: []simplejson.TableQueryResponseColumn{
			{Text: "timestamp", Data: simplejson.TableQueryResponseTimeColumn(timestamps)},
			{Text: "country", Data: simplejson.TableQueryResponseStringColumn(countryNames)},
			{Text: "???", Data: simplejson.TableQueryResponseNumberColumn(values)},
		},
	}
}
