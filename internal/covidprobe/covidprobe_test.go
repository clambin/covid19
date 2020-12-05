package covidprobe

import (
	"testing"
	"github.com/stretchr/testify/assert"

	"covid19/internal/coviddb/mock"
)

// whitebox testing of CovidProbe
func TestCovidProbeWhite(t *testing.T) {
	apiClient := makeClient()
	db        := mock.Create(testDBData)
	probe     := NewCovidProbe(apiClient, db, nil)

	countryStats, err := probe.apiClient.GetCountryStats()
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(countryStats))
	newRecords, err := probe.findNewCountryStats(countryStats)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(newRecords))

	// go doesn't guarantee we will get our records in the expected order
	indices := []struct{ name string; index int}{
		{ name: "A", index: 0, },
		{ name: "B", index: 1, }}

	if newRecords[0].Name == "B" {
		indices[0].index = 1
		indices[1].index = 0
	}

	assert.Equal(t, true,     newRecords[indices[0].index].Timestamp.Equal(lastUpdate))
	assert.Equal(t, "A",      newRecords[indices[0].index].Name)
	assert.Equal(t, "AA",     newRecords[indices[0].index].Code)
	assert.Equal(t, int64(3), newRecords[indices[0].index].Confirmed)
	assert.Equal(t, int64(2), newRecords[indices[0].index].Deaths)
	assert.Equal(t, int64(1), newRecords[indices[0].index].Recovered)

	assert.Equal(t, true,     newRecords[indices[1].index].Timestamp.Equal(lastUpdate))
	assert.Equal(t, "B",      newRecords[indices[1].index].Name)
	assert.Equal(t, "BB",     newRecords[indices[1].index].Code)
	assert.Equal(t, int64(6), newRecords[indices[1].index].Confirmed)
	assert.Equal(t, int64(5), newRecords[indices[1].index].Deaths)
	assert.Equal(t, int64(4), newRecords[indices[1].index].Recovered)

	return
}

// blackbox testing of CovidProbe
func TestCovidProbeBlack(t *testing.T) {
	apiClient := makeClient()
	db        := mock.Create(testDBData)
	probe     := NewCovidProbe(apiClient, db, nil)

	err := probe.Run()

	assert.Equal(t, nil, err)
	latest, err := db.ListLatestByCountry()
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(latest))
	assert.Equal(t, true, latest["A"].Equal(lastUpdate))
	assert.Equal(t, true, latest["B"].Equal(lastUpdate))
}

