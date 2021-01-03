package reporters_test

import (
	"testing"
	"time"

	"github.com/clambin/gotools/metrics"
	"github.com/stretchr/testify/assert"

	"covid19/internal/coviddb"
	"covid19/internal/reporters"
)

func TestCountriesReporter(t *testing.T) {
	r := reporters.Create()
	r.Add(reporters.NewCountriesReporter("http://localhost:8080"))

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

	var (
		value float64
		err   error
	)
	value, err = metrics.LoadValue("covid_reported_count", "Belgium")
	assert.Nil(t, err)
	assert.Equal(t, 1.0, value)
	value, err = metrics.LoadValue("covid_reported_count", "US")
	assert.Nil(t, err)
	assert.Equal(t, 1.0, value)
}
