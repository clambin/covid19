package utils

import (
	"sort"
	"time"
)

func GetUniqueTimestamps[T any](timeMap map[time.Time]T) (timestamps []time.Time) {
	timestamps = make([]time.Time, 0, len(timeMap))
	for timestamp := range timeMap {
		timestamps = append(timestamps, timestamp)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i].Before(timestamps[j]) })
	return
}
