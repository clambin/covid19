package probe_test

import (
	coviddb2 "github.com/clambin/covid19/coviddb"
	"github.com/clambin/covid19/coviddb/mock"
	mock2 "github.com/clambin/covid19/population/db/mock"
	probe2 "github.com/clambin/covid19/population/probe"
	"github.com/clambin/gotools/httpstub"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPopulationProbe(t *testing.T) {
	covidEntries := []coviddb2.CountryEntry{
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
	covidDB := mock.Create(covidEntries)
	popDB := mock2.Create(map[string]int64{})

	p := probe2.Create("", popDB, covidDB)
	p.APIClient = &probe2.RapidAPIClient{
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
