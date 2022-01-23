package updates

import (
	"context"
	covidStore "github.com/clambin/covid19/covid/store"
	"github.com/clambin/simplejson"
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
		TableQuery: handler.tableQuery,
	}
}

func (handler *Handler) tableQuery(_ context.Context, args *simplejson.TableQueryArgs) (response *simplejson.TableQueryResponse, err error) {
	var entries map[time.Time]int
	entries, err = handler.CovidDB.CountEntriesByTime(args.Range.From, args.Range.To)
	if err != nil {
		return
	}

	timestamps := getUniqueSortedTimestamps(entries)
	var updateCount []float64
	for _, timestamp := range timestamps {
		updateCount = append(updateCount, float64(entries[timestamp]))
	}

	return &simplejson.TableQueryResponse{
		Columns: []simplejson.TableQueryResponseColumn{
			{Text: "timestamp", Data: simplejson.TableQueryResponseTimeColumn(timestamps)},
			{Text: "updates", Data: simplejson.TableQueryResponseNumberColumn(updateCount)},
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
