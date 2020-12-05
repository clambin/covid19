package covidprobe

import (
	"testing"
	"github.com/stretchr/testify/assert"

	// "covid19/internal/coviddb"
	"covid19/internal/coviddb/mock"
)

// whitebox testing of CovidProbe
func TestCovidProbeWhite(t *testing.T) {
	apiClient := makeClient()
	db        := mock.Create(testDBData)
	probe     := NewCovidProbe(apiClient, db, "")

	countryStats, err := probe.apiClient.GetCountryStats()
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(countryStats))
	newRecords, err := probe.findNewCountryStats(countryStats)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(newRecords))
	assert.Equal(t, true, newRecords[0].Timestamp.Equal(lastUpdate))
	assert.Equal(t, "A", newRecords[0].Name)
	assert.Equal(t, "AA", newRecords[0].Code)
	assert.Equal(t, int64(3), newRecords[0].Confirmed)
	assert.Equal(t, int64(2), newRecords[0].Deaths)
	assert.Equal(t, int64(1), newRecords[0].Recovered)
	assert.Equal(t, true, newRecords[1].Timestamp.Equal(lastUpdate))
	assert.Equal(t, "B", newRecords[1].Name)
	assert.Equal(t, "BB", newRecords[1].Code)
	assert.Equal(t, int64(6), newRecords[1].Confirmed)
	assert.Equal(t, int64(5), newRecords[1].Deaths)
	assert.Equal(t, int64(4), newRecords[1].Recovered)

	return
}

// blackbox testing of CovidProbe
func TestCovidProbeBlack(t *testing.T) {
	apiClient := makeClient()
	db        := mock.Create(testDBData)
	probe     := NewCovidProbe(apiClient, db, "")

	err := probe.Run()

	assert.Equal(t, nil, err)
	latest, err := db.ListLatestByCountry()
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(latest))
	assert.Equal(t, true, latest["A"].Equal(lastUpdate))
	assert.Equal(t, true, latest["B"].Equal(lastUpdate))
}

