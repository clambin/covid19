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
	newDataReporter := reporters.NewUpdatesReporter("", "", []string{"Belgium"}, db)
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
	assert.Equal(t, "New covid19 data", newDataReporter.SentReqs[0].Title)
	assert.Equal(t, "New data for Belgium. New confirmed: 5. New deaths: 0. New recovered: 1", newDataReporter.SentReqs[0].Message)
}
