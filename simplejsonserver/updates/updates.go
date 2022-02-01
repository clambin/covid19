package updates

import (
	"context"
	covidStore "github.com/clambin/covid19/covid/store"
	"github.com/clambin/simplejson/v3"
	"github.com/clambin/simplejson/v3/query"
	"sort"
	"time"
)

// Handler calculates the number of updated countries by timestamp
type Handler struct {
	CovidDB covidStore.CovidStore
}

var _ simplejson.Handler = &Handler{}

func (handler Handler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		Query: handler.tableQuery,
	}
}

func (handler *Handler) tableQuery(_ context.Context, req query.Request) (response query.Response, err error) {
	var entries map[time.Time]int
	entries, err = handler.CovidDB.CountEntriesByTime(req.Args.Range.From, req.Args.Range.To)
	if err != nil {
		return
	}

	timestamps := getUniqueSortedTimestamps(entries)
	var updateCount []float64
	for _, timestamp := range timestamps {
		updateCount = append(updateCount, float64(entries[timestamp]))
	}

	return &query.TableResponse{
		Columns: []query.Column{
			{Text: "timestamp", Data: query.TimeColumn(timestamps)},
			{Text: "updates", Data: query.NumberColumn(updateCount)},
		},
	}, nil
}

func getUniqueSortedTimestamps(updates map[time.Time]int) (timestamps []time.Time) {
	for key := range updates {
		timestamps = append(timestamps, key)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i].Before(timestamps[j]) })
	return
}
