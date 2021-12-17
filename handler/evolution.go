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
		values = append(values, float64(increases[name]))
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
	entries, err = handler.Cache.DB.GetAll()

	if err != nil {
		return
	}

	confirmed = make(map[string][]int64)

	var endTime time.Time
	for index := len(entries) - 1; index >= 0; index-- {
		if entries[index].Timestamp.After(end) {
			continue
		}
		if endTime.IsZero() {
			endTime = entries[index].Timestamp
		}
		delta := endTime.Sub(entries[index].Timestamp)
		if delta > evolutionWindow*24*time.Hour {
			break
		}

		list, _ := confirmed[entries[index].Code]
		list = append(list, entries[index].Confirmed)
		confirmed[entries[index].Code] = list
	}

	return
}

func getIncreases(confirmed map[string][]int64) (increases map[string]int64) {
	increases = make(map[string]int64)
	for key, list := range confirmed {
		var increase int64
		if len(list) > 0 {
			increase = list[0] - list[len(list)-1]
		}
		increases[key] = increase
	}
	return
}

func getSortedCountryNames(averages map[string]int64) (names []string) {
	for name := range averages {
		names = append(names, name)
	}
	sort.Strings(names)
	return
}
