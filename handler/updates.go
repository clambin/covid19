package handler

import (
	"github.com/clambin/covid19/models"
	grafana_json "github.com/clambin/grafana-json"
	"sort"
	"time"
)

func (handler *Handler) handleUpdates(args *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
	var entries []models.CountryEntry
	entries, err = handler.Cache.DB.GetAll()
	if err != nil {
		return
	}

	updates := make(map[time.Time]int)

	for _, entry := range entries {
		if args.Range.To.IsZero() == false && entry.Timestamp.After(args.Range.To) {
			break
		}

		count, _ := updates[entry.Timestamp]
		count++
		updates[entry.Timestamp] = count
	}

	timestamps := getUniqueSortedTimestamps(updates)
	var updateCount []float64
	for _, timestamp := range timestamps {
		updateCount = append(updateCount, float64(updates[timestamp]))
	}

	return &grafana_json.TableQueryResponse{
		Columns: []grafana_json.TableQueryResponseColumn{
			{Text: "timestamp", Data: grafana_json.TableQueryResponseTimeColumn(timestamps)},
			{Text: "updates", Data: grafana_json.TableQueryResponseNumberColumn(updateCount)},
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
