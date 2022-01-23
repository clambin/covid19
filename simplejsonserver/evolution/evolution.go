package evolution

import (
	"context"
	covidStore "github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson"
	"sort"
	"time"
)

// Window is the number of past days that will be used to calculate the average
const Window = 7

// Handler calculates the 7-day average increase in confirmed cases by country
type Handler struct {
	CovidDB covidStore.CovidStore
}

var _ simplejson.Handler = &Handler{}

func (handler Handler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		TableQuery: handler.tableQuery,
	}
}

func (handler *Handler) tableQuery(_ context.Context, args *simplejson.TableQueryArgs) (response *simplejson.TableQueryResponse, err error) {
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

	return &simplejson.TableQueryResponse{
		Columns: []simplejson.TableQueryResponseColumn{
			{Text: "timestamp", Data: simplejson.TableQueryResponseTimeColumn(timestamps)},
			{Text: "country", Data: simplejson.TableQueryResponseStringColumn(names)},
			{Text: "increase", Data: simplejson.TableQueryResponseNumberColumn(values)},
		},
	}, nil
}

func (handler *Handler) getLatestEntries(end time.Time) (confirmed map[string][]int64, err error) {
	var entries []models.CountryEntry

	if end.IsZero() {
		entries, err = handler.CovidDB.GetAll()
	} else {
		entries, err = handler.CovidDB.GetAllForRange(end.Add(-Window*24*time.Hour), end)
	}

	if err != nil || len(entries) == 0 {
		return
	}

	if end.IsZero() {
		start := entries[len(entries)-1].Timestamp.Add(-Window * 24 * time.Hour)

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
