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
		{time.Now().AddDate(0, -1, 0), "BE", "Belgium", 5, 1, 0},
	})
	r := reporters.Create()
	newDataReporter := reporters.NewUpdatesReporter("", "", []string{"Belgium"}, db)
	r.Add(newDataReporter)

	r.Report([]coviddb.CountryEntry{
		{time.Now(), "BE", "Belgium", 10, 1, 1},
		{time.Now(), "US", "US", 10, 1, 1},
	})

	assert.Len(t, newDataReporter.SentReqs, 1)
	assert.Equal(t, "New covid19 data", newDataReporter.SentReqs[0].Title)
	assert.Equal(t, "New data for Belgium. New confirmed: 5. New deaths: 1. New recovered: 0", newDataReporter.SentReqs[0].Message)
}
