package handler

import (
	"github.com/clambin/covid19/models"
	grafana_json "github.com/clambin/grafana-json"
	"sort"
	"time"
)

const evolutionWindow = 7

func (handler *Handler) handleEvolution(args *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
	var entries map[string][]int64
	entries, err = handler.getLatestEntries(args.Range.To)
	if err != nil {
		return
	}

	increases := getIncreases(entries)
	names := getSortedCountryNames(increases)

	var (
		timestamps []time.Time
		values     []float64
	)

	for _, name := range names {
		timestamps = append(timestamps, args.Range.To)
		values = append(values, increases[name])
	}

	return &grafana_json.TableQueryResponse{
		Columns: []grafana_json.TableQueryResponseColumn{
			{Text: "timestamp", Data: grafana_json.TableQueryResponseTimeColumn(timestamps)},
			{Text: "country", Data: grafana_json.TableQueryResponseStringColumn(names)},
			{Text: "increase", Data: grafana_json.TableQueryResponseNumberColumn(values)},
		},
	}, nil
}

func (handler *Handler) getLatestEntries(end time.Time) (confirmed map[string][]int64, err error) {
	var entries []models.CountryEntry

	if end.IsZero() {
		entries, err = handler.Cache.DB.GetAll()
	} else {
		entries, err = handler.Cache.DB.GetAllForRange(end.Add(-evolutionWindow*24*time.Hour), end)
	}

	if err != nil || len(entries) == 0 {
		return
	}

	if end.IsZero() {
		start := entries[len(entries)-1].Timestamp.Add(-evolutionWindow * 24 * time.Hour)

		for len(entries) > 0 && entries[0].Timestamp.Before(start) {
			entries = entries[1:]
		}
	}

	confirmed = make(map[string][]int64)

	for index := len(entries) - 1; index >= 0; index-- {
		list, _ := confirmed[entries[index].Code]
		list = append(list, entries[index].Confirmed)
		confirmed[entries[index].Code] = list
	}

	return
}

func getIncreases(confirmed map[string][]int64) (increases map[string]float64) {
	increases = make(map[string]float64)
	for key, list := range confirmed {
		var increase float64
		count := len(list)
		if count > 0 {
			increase = float64(list[0]-list[len(list)-1]) / float64(count)
		}
		increases[key] = increase
	}
	return
}

func getSortedCountryNames(averages map[string]float64) (names []string) {
	for name := range averages {
		names = append(names, name)
	}
	sort.Strings(names)
	return
}

func (handler *Handler) handleMortalityVsConfirmed(args *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
	var countryNames []string
	countryNames, err = handler.Cache.DB.GetAllCountryNames()
	if err != nil {
		return
	}

	var entries map[string]models.CountryEntry
	entries, err = handler.Cache.DB.GetLatestForCountriesByTime(countryNames, args.Range.To)
	if err != nil {
		return
	}

	var timestamps []time.Time
	var countryCodes []string
	var ratios []float64

	for _, countryName := range countryNames {
		entry, ok := entries[countryName]
		if ok == false {
			continue
		}

		timestamps = append(timestamps, entry.Timestamp)
		countryCodes = append(countryCodes, entry.Code)
		var ratio float64
		if entry.Confirmed > 0 {
			ratio = float64(entry.Deaths) / float64(entry.Confirmed)
		}
		ratios = append(ratios, ratio)
	}

	return &grafana_json.TableQueryResponse{Columns: []grafana_json.TableQueryResponseColumn{
		{
			Text: "timestamp",
			Data: grafana_json.TableQueryResponseTimeColumn(timestamps),
		},
		{
			Text: "country",
			Data: grafana_json.TableQueryResponseStringColumn(countryCodes),
		},
		{
			Text: "ratio",
			Data: grafana_json.TableQueryResponseNumberColumn(ratios),
		},
	}}, nil
}
