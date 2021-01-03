package monitor_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"covid19/internal/covid/apiclient"
	mockapi1 "covid19/internal/covid/apiclient/mock"
	"covid19/internal/covid/probe"
	"covid19/internal/coviddb"
	mockdb1 "covid19/internal/coviddb/mock"
	"covid19/internal/monitor"
	mockapi2 "covid19/internal/population/apiclient/mock"
	mockdb2 "covid19/internal/population/db/mock"
	popprobe "covid19/internal/population/probe"
	"covid19/internal/pushgateway"
)

func TestMonitor(t *testing.T) {
	cfg := monitor.Configuration{Once: true, Debug: true}

	db1 := mockdb1.Create([]coviddb.CountryEntry{})
	api1 := mockapi1.New(map[string]apiclient.CountryStats{
		"Belgium":     {LastUpdate: time.Now(), Confirmed: 4, Recovered: 2, Deaths: 1},
		"US":          {LastUpdate: time.Now(), Confirmed: 20, Recovered: 15, Deaths: 5},
		"NotACountry": {LastUpdate: time.Now(), Confirmed: 0, Recovered: 0, Deaths: 0},
	})
	covidProbe := probe.NewProbe(api1, db1, pushgateway.NewPushGateway("localhost:8080"))

	db2 := mockdb2.Create(map[string]int64{})
	api2 := mockapi2.New(map[string]int64{
		"BE": int64(11248330),
		"US": int64(321645000),
	})
	popProbe := popprobe.Create(api2, db2)

	ok := monitor.Run(&cfg, covidProbe, popProbe)
	assert.True(t, ok)

	covidEntries, err := db1.List(time.Now())
	assert.Nil(t, err)
	assert.Len(t, covidEntries, 2)

	popEntries, err := db2.List()
	assert.Nil(t, err)
	assert.Len(t, popEntries, 2)
}
