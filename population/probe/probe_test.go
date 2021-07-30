package probe_test

import (
	"context"
	"github.com/clambin/covid19/coviddb"
	covidDBMock "github.com/clambin/covid19/coviddb/mock"
	popDBMock "github.com/clambin/covid19/population/db/mock"
	"github.com/clambin/covid19/population/probe"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
			Code:      "??",
			Confirmed: 10,
		},
	}
	covidDB := covidDBMock.Create(covidEntries)
	popDB := popDBMock.Create(map[string]int64{})

	server := httptest.NewServer(http.HandlerFunc(serverStub))
	defer server.Close()

	p := probe.Create("1234", popDB, covidDB)
	p.APIClient.(*probe.RapidAPIClient).Client.URL = server.URL

	// DB should be empty
	entries, err := popDB.List()
	assert.NoError(t, err)
	assert.Len(t, entries, 0)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run the probe
	go func() {
		err = p.Run(ctx, 50*time.Millisecond)
		assert.NoError(t, err)
	}()

	for i := 0; i < 2; i++ {
		assert.Eventually(t, func() bool {
			checkEntries, err2 := popDB.List()
			return err2 == nil && len(checkEntries) == 2
		}, 500*time.Millisecond, 10*time.Millisecond)

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

		popDB.DeleteAll()
	}
}
