package handler

import (
	"github.com/clambin/covid19/covid/probe/fetcher"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson"
	"time"
)

const (
	countryConfirmed = iota
	countryDeaths
)

func (handler *Handler) handleLatestConfirmedByCountry(args *simplejson.TableQueryArgs) (response *simplejson.TableQueryResponse, err error) {
	return handler.handleLatestCountryStats(args.Range, countryConfirmed)
}

func (handler *Handler) handleLatestDeathsByCountry(args *simplejson.TableQueryArgs) (response *simplejson.TableQueryResponse, err error) {
	return handler.handleLatestCountryStats(args.Range, countryDeaths)
}

func (handler *Handler) handleLatestCountryStats(args simplejson.Range, mode int) (response *simplejson.TableQueryResponse, err error) {
	var names []string
	names, err = handler.Cache.DB.GetAllCountryNames()

	if err != nil {
		return
	}

	var entries map[string]models.CountryEntry
	if args.To.IsZero() {
		entries, err = handler.Cache.DB.GetLatestForCountries(names)
	} else {
		entries, err = handler.Cache.DB.GetLatestForCountriesByTime(names, args.To)
	}

	if err != nil {
		return
	}

	var timestamp time.Time
	response = &simplejson.TableQueryResponse{}

	for _, name := range names {
		entry := entries[name]
		if timestamp.IsZero() {
			timestamp = entry.Timestamp
			response.Columns = append(response.Columns, simplejson.TableQueryResponseColumn{
				Text: "timestamp",
				Data: simplejson.TableQueryResponseTimeColumn([]time.Time{timestamp}),
			})
		}

		var value float64
		switch mode {
		case countryConfirmed:
			value = float64(entry.Confirmed)
		case countryDeaths:
			value = float64(entry.Deaths)
		}

		response.Columns = append(response.Columns, simplejson.TableQueryResponseColumn{
			Text: name,
			Data: simplejson.TableQueryResponseNumberColumn([]float64{value}),
		})
	}

	return
}

func (handler *Handler) handleConfirmedByCountryByPopulation(args *simplejson.TableQueryArgs) (response *simplejson.TableQueryResponse, err error) {
	return handler.handleCountryStatsByPopulation(args, countryConfirmed)
}

func (handler *Handler) handleDeathsByCountryByPopulation(args *simplejson.TableQueryArgs) (response *simplejson.TableQueryResponse, err error) {
	return handler.handleCountryStatsByPopulation(args, countryDeaths)
}

func (handler *Handler) handleCountryStatsByPopulation(args *simplejson.TableQueryArgs, mode int) (response *simplejson.TableQueryResponse, err error) {
	if response, err = handler.handleLatestCountryStats(args.Range, mode); err != nil {
		return
	}

	response = handler.pivotResponse(response)

	var population map[string]int64
	if population, err = handler.PopulationStore.List(); err != nil {
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
	switch mode {
	case countryConfirmed:
		response.Columns[2].Text = "confirmed"
	case countryDeaths:
		response.Columns[2].Text = "deaths"
	}

	return
}

func (handler *Handler) pivotResponse(input *simplejson.TableQueryResponse) (output *simplejson.TableQueryResponse) {
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
