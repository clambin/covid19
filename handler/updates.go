package handler

import (
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson"
	"sort"
	"time"
)

func (handler *Handler) handleUpdates(args *simplejson.TableQueryArgs) (response *simplejson.TableQueryResponse, err error) {
	var entries []models.CountryEntry
	entries, err = handler.Cache.DB.GetAll()
	if err != nil {
		return
	}

	last := make(map[string]models.CountryEntry)
	updates := make(map[time.Time]int)

	for _, entry := range entries {
		if args.Range.To.IsZero() == false && entry.Timestamp.After(args.Range.To) {
			break
		}

		previous, _ := last[entry.Code]
		if entry.Confirmed != previous.Confirmed || entry.Deaths != previous.Deaths || entry.Recovered != previous.Recovered {
			count, _ := updates[entry.Timestamp]
			count++
			updates[entry.Timestamp] = count
		}
		last[entry.Code] = entry
	}

	timestamps := getUniqueSortedTimestamps(updates)
	var updateCount []float64
	for _, timestamp := range timestamps {
		updateCount = append(updateCount, float64(updates[timestamp]))
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
