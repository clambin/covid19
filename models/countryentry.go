package models

import "time"

// CountryEntry represents one entry of COVID-19 statistics
type CountryEntry struct {
	Timestamp time.Time
	Code      string
	Name      string
	Confirmed int64
	Recovered int64
	Deaths    int64
}
