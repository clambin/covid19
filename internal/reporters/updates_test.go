package reporters_test

import (
	mockdb "covid19/internal/coviddb/mock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"covid19/internal/coviddb"
	"covid19/internal/reporters"
)

func TestNewDataReporter(t *testing.T) {
	db := mockdb.Create([]coviddb.CountryEntry{
		{
			Timestamp: time.Now().AddDate(0, -1, 0),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 5,
			Deaths:    1,
			Recovered: 0,
		},
	})
	r := reporters.Create()

	reportersConfig := reporters.ReportsConfiguration{}
	reportersConfig.Countries = []string{"Belgium"}
	reportersConfig.Updates.Pushover.Token = "1234"
	reportersConfig.Updates.Pushover.Token = "5678"

	newDataReporter := reporters.NewUpdatesReporter(&reportersConfig, db)

	r.Add(newDataReporter)

	r.Report([]coviddb.CountryEntry{
		{
			Timestamp: time.Now(),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 10,
			Deaths:    1,
			Recovered: 1,
		},
		{
			Timestamp: time.Now(),
			Code:      "US",
			Name:      "US",
			Confirmed: 10,
			Deaths:    1,
			Recovered: 1,
		},
	})

	assert.Len(t, newDataReporter.SentReqs, 1)
	assert.Equal(t, "New confirmed: 5\nNew deaths: 0\nNew recovered: 1", newDataReporter.SentReqs[0])
}
