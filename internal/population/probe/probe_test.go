package probe_test

import (
	"github.com/clambin/covid19/internal/coviddb"
	mockCovidDB "github.com/clambin/covid19/internal/coviddb/mock"
	mockPopDB "github.com/clambin/covid19/internal/population/db/mock"
	"github.com/clambin/covid19/internal/population/probe"
	"github.com/clambin/gotools/httpstub"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPopulationProbe(t *testing.T) {
	covidEntries := []coviddb.CountryEntry{
		{
			Code:      "BE",
			Confirmed: 100,
		},
		{
			Code:      "US",
			Confirmed: 300,
		},
		{
			Code:      "FO",
			Confirmed: 10,
		},
	}
	covidDB := mockCovidDB.Create(covidEntries)
	popDB := mockPopDB.Create(map[string]int64{})

	p := probe.Create("", popDB, covidDB)
	p.APIClient = &probe.RapidAPIClient{
		HTTPClient: httpstub.NewTestClient(serverStub),
		APIKey:     "1234",
	}

	// DB should be empty
	entries, err := popDB.List()
	assert.NoError(t, err)
	assert.Len(t, entries, 0)

	// Run the probe
	err = p.Run()
	assert.NoError(t, err)

	// Check results
	entries, err = popDB.List()
	assert.NoError(t, err)
	assert.Len(t, entries, 2)

	val, ok := entries["BE"]
	assert.True(t, ok)
	assert.Equal(t, int64(20), val)

	val, ok = entries["US"]
	assert.True(t, ok)
	assert.Equal(t, int64(40), val)
}
