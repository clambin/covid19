package handler

import (
	"github.com/clambin/covid19/covid/probe/fetcher"
	"github.com/clambin/covid19/models"
	grafana_json "github.com/clambin/grafana-json"
	"time"
)

const (
	countryConfirmed = iota
	countryDeaths
)

func (handler *Handler) handleLatestConfirmedByCountry(args *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
	return handler.handleLatestCountryStats(args.Range, countryConfirmed)
}

func (handler *Handler) handleLatestDeathsByCountry(args *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
	return handler.handleLatestCountryStats(args.Range, countryDeaths)
}

func (handler *Handler) handleLatestCountryStats(args grafana_json.QueryRequestRange, mode int) (response *grafana_json.TableQueryResponse, err error) {
	var names []string
	names, err = handler.Cache.DB.GetAllCountryNames()

	if err != nil {
		return
	}

	var entries map[string]*models.CountryEntry
	if args.To.IsZero() {
		entries, err = handler.Cache.DB.GetLatestForCountries(names)
	} else {
		entries, err = handler.Cache.DB.GetLatestForCountriesByTime(names, args.To)
	}

	if err != nil {
		return
	}

	var timestamp time.Time
	response = &grafana_json.TableQueryResponse{}

	for _, name := range names {
		entry := entries[name]
		if timestamp.IsZero() {
			timestamp = entry.Timestamp
			response.Columns = append(response.Columns, grafana_json.TableQueryResponseColumn{
				Text: "timestamp",
				Data: grafana_json.TableQueryResponseTimeColumn([]time.Time{timestamp}),
			})
		}

		var value float64
		switch mode {
		case countryConfirmed:
			value = float64(entry.Confirmed)
		case countryDeaths:
			value = float64(entry.Deaths)
		}

		response.Columns = append(response.Columns, grafana_json.TableQueryResponseColumn{
			Text: name,
			Data: grafana_json.TableQueryResponseNumberColumn([]float64{value}),
		})
	}

	return
}

func (handler *Handler) handleConfirmedByCountryByPopulation(args *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
	return handler.handleCountryStatsByPopulation(args, countryConfirmed)
}

func (handler *Handler) handleDeathsByCountryByPopulation(args *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
	return handler.handleCountryStatsByPopulation(args, countryDeaths)
}

func (handler *Handler) handleCountryStatsByPopulation(args *grafana_json.TableQueryArgs, mode int) (response *grafana_json.TableQueryResponse, err error) {
	if response, err = handler.handleLatestCountryStats(args.Range, mode); err != nil {
		return
	}

	response = handler.pivotResponse(response)

	var population map[string]int64
	if population, err = handler.PopulationStore.List(); err != nil {
		return
	}

	// calculate figure by country population
	for index := range response.Columns[0].Data.(grafana_json.TableQueryResponseTimeColumn) {
		name := response.Columns[1].Data.(grafana_json.TableQueryResponseStringColumn)[index]
		count := response.Columns[2].Data.(grafana_json.TableQueryResponseNumberColumn)[index]

		code, codeFound := fetcher.CountryCodes[name]
		if codeFound == false {
			code = name
		}
		response.Columns[1].Data.(grafana_json.TableQueryResponseStringColumn)[index] = code

		var rate float64
		if pop, popFound := population[code]; popFound {
			rate = count / float64(pop)
		}
		response.Columns[2].Data.(grafana_json.TableQueryResponseNumberColumn)[index] = rate
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

func (handler *Handler) pivotResponse(input *grafana_json.TableQueryResponse) (output *grafana_json.TableQueryResponse) {
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

	timestamp := input.Columns[0].Data.(grafana_json.TableQueryResponseTimeColumn)[0]

	for _, col := range input.Columns[1:] {
		countryName := col.Text
		value := col.Data.(grafana_json.TableQueryResponseNumberColumn)[0]

		timestamps = append(timestamps, timestamp)
		countryNames = append(countryNames, countryName)
		values = append(values, value)
	}

	return &grafana_json.TableQueryResponse{
		Columns: []grafana_json.TableQueryResponseColumn{
			{Text: "timestamp", Data: grafana_json.TableQueryResponseTimeColumn(timestamps)},
			{Text: "country", Data: grafana_json.TableQueryResponseStringColumn(countryNames)},
			{Text: "???", Data: grafana_json.TableQueryResponseNumberColumn(values)},
		},
	}
}
