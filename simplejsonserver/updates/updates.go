package updates

import (
	"context"
	"github.com/clambin/covid19/db"
	"github.com/clambin/simplejson/v6"
	"github.com/clambin/simplejson/v6/pkg/data"
	"time"
)

// Handler calculates the number of updated countries by timestamp
type Handler struct {
	DB CovidGetter
}

type CovidGetter interface {
	CountEntriesByTime(time.Time, time.Time) ([]db.TimestampCount, error)
}

var _ simplejson.Handler = &Handler{}

func (handler *Handler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		Query: handler.tableQuery,
	}
}

func (handler *Handler) tableQuery(_ context.Context, req simplejson.QueryRequest) (simplejson.Response, error) {
	entries, err := handler.DB.CountEntriesByTime(req.Args.Range.From, req.Args.Range.To)
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
