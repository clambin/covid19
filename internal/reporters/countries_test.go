package reporters_test

import (
	"covid19/internal/coviddb"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/clambin/gotools/metrics"

	"covid19/internal/reporters"
)

func TestCountriesReporter(t *testing.T) {
	r := reporters.Create()
	r.Add(reporters.NewCountriesReporter("http://localhost:8080"))

	r.Report([]coviddb.CountryEntry{
		{time.Now(), "BE", "Belgium", 10, 1, 1},
		{time.Now(), "US", "US", 10, 1, 1},
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
