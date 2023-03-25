package evolution

import (
	"context"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v6"
	"sort"
	"time"
)

// Window is the number of past days that will be used to calculate the average
const Window = 7

// Handler calculates the 7-day average increase in confirmed cases by country
type Handler struct {
	CovidDB CovidGetter
}

type CovidGetter interface {
	GetAllForRange(time.Time, time.Time) ([]models.CountryEntry, error)
}

var _ simplejson.Handler = &Handler{}

func (handler *Handler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		Query: handler.tableQuery,
	}
}

func (handler *Handler) tableQuery(_ context.Context, req simplejson.QueryRequest) (simplejson.Response, error) {
	end := req.Args.Range.To
	if end.IsZero() {
		end = time.Now()
	}

	entries, err := handler.CovidDB.GetAllForRange(end.Add(-Window*24*time.Hour), end)
	if err != nil {
		return nil, err
	}

	var (
		timestamps []time.Time
		names      []string
		values     []float64
	)

	if len(entries) > 0 {
		start := entries[len(entries)-1].Timestamp.Add(-Window * 24 * time.Hour)

		increases := processEntries(entries, start)
		names = getSortedCountryNames(increases)

		timestamps = make([]time.Time, len(names))
		values = make([]float64, len(names))

		for idx, name := range names {
			// TODO: if To is zero, all reported timestamps are zero
			timestamps[idx] = req.Args.Range.To
			values[idx] = increases[name]
		}
	}

	return &simplejson.TableResponse{
		Columns: []simplejson.Column{
			{Text: "timestamp", Data: simplejson.TimeColumn(timestamps)},
			{Text: "country", Data: simplejson.StringColumn(names)},
			{Text: "increase", Data: simplejson.NumberColumn(values)},
		},
	}, nil
}

func getSortedCountryNames(averages map[string]float64) (names []string) {
	for name := range averages {
		names = append(names, name)
	}
	sort.Strings(names)
	return
}

func processEntries(entries []models.CountryEntry, start time.Time) (output map[string]float64) {
	summary := summarizeEntries(entries, start)

	output = make(map[string]float64)
	for key, entry := range summary {
		output[key] = entry.increase()
	}
	return
}

func summarizeEntries(entries []models.CountryEntry, start time.Time) (summary map[string]*evolution) {
	summary = make(map[string]*evolution)
	for i := 0; i < len(entries); i++ {
		if entries[i].Timestamp.Before(start) {
			continue
		}
		current, ok := summary[entries[i].Code]
		if !ok {
			current = &evolution{}
			summary[entries[i].Code] = current
		}
		current.process(entries[i])
		//summary[entries[i].Code] = current
	}
	return
}

type evolution struct {
	first evolutionEntry
	last  evolutionEntry
}

func (e *evolution) process(entry models.CountryEntry) {
	if e.first.timestamp.IsZero() {
		e.first.timestamp = entry.Timestamp
		e.first.value = entry.Confirmed
	} else {
		e.last.timestamp = entry.Timestamp
		e.last.value = entry.Confirmed
	}
}

func (e *evolution) increase() float64 {
	if e.first.timestamp.IsZero() || e.last.timestamp.IsZero() {
		return 0
	}
	days := float64(e.last.timestamp.Sub(e.first.timestamp).Hours()) / 24.0

	return float64(e.last.value-e.first.value) / days
}

type evolutionEntry struct {
	timestamp time.Time
	value     int64
}
