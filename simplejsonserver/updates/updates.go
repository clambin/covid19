package updates

import (
	"context"
	covidStore "github.com/clambin/covid19/covid/store"
	"github.com/clambin/simplejson/v3"
	"github.com/clambin/simplejson/v3/dataset"
	"github.com/clambin/simplejson/v3/query"
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

	d := dataset.New()
	for timestamp, count := range entries {
		d.Add(timestamp, "updates", float64(count))
	}

	return d.GenerateTableResponse(), nil
}
