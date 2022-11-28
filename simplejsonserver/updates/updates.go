package updates

import (
	"context"
	covidStore "github.com/clambin/covid19/db"
	"github.com/clambin/simplejson/v4"
	"github.com/clambin/simplejson/v4/pkg/data"
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

func (handler *Handler) tableQuery(_ context.Context, req simplejson.QueryRequest) (simplejson.Response, error) {
	// TODO: have CountEntriesByTime return a (sorted) slice of timestamp/count pairs so we don't have to sort here.
	entries, err := handler.CovidDB.CountEntriesByTime(req.Args.Range.From, req.Args.Range.To)
	if err != nil {
		return nil, err
	}

	var timestamps []time.Time
	var counts []float64

	for _, entry := range entries {
		timestamps = append(timestamps, entry.Timestamp)
		counts = append(counts, float64(entry.Count))
	}

	return data.New(
		data.Column{Name: "timestamp", Values: timestamps},
		data.Column{Name: "updates", Values: counts},
	).CreateTableResponse(), nil
}
