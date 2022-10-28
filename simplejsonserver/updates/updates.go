package updates

import (
	"context"
	covidStore "github.com/clambin/covid19/db"
	"github.com/clambin/simplejson/v3"
	"github.com/clambin/simplejson/v3/data"
	"github.com/clambin/simplejson/v3/query"
	"sort"
	"time"
)

// Handler calculates the number of updated countries by timestamp
type Handler struct {
	CovidDB covidStore.CovidStore
}

var _ simplejson.Handler = &Handler{}

func (handler *Handler) Endpoints() (endpoints simplejson.Endpoints) {
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

	timestamps, updates := getSortedUpdates(entries)
	d := data.New(
		data.Column{Name: "timestamp", Values: timestamps},
		data.Column{Name: "updates", Values: updates},
	)

	return d.CreateTableResponse(), nil
}

func getSortedUpdates(entries map[time.Time]int) ([]time.Time, []float64) {
	type updateEntry struct {
		timestamp time.Time
		updates   float64
	}
	var result []updateEntry
	for timestamp, update := range entries {
		result = append(result, updateEntry{
			timestamp: timestamp,
			updates:   float64(update),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].timestamp.Before(result[j].timestamp)
	})

	var timestamps []time.Time
	var updates []float64
	for _, entry := range result {
		timestamps = append(timestamps, entry.timestamp)
		updates = append(updates, entry.updates)
	}
	return timestamps, updates
}
