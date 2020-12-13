package db_test

import (
	"os"
	"strconv"
	"time"

	"github.com/stretchr/testify/assert"
	"testing"

	"covid19/internal/covid/db"
)

func getDBEnv() (map[string]string, bool) {
	values := make(map[string]string, 0)
	envVars := []string{"pg_host", "pg_port", "pg_database", "pg_user", "pg_password"}

	ok := true
	for _, envVar := range envVars {
		value, found := os.LookupEnv(envVar)
		if found {
			values[envVar] = value
		} else {
			ok = false
			break
		}
	}

	return values, ok
}

func TestDB(t *testing.T) {
	values, ok := getDBEnv()
	if ok == false {
		t.Log("Could not find all DB env variables. Skipping this test")
		return
	}

	port, err := strconv.Atoi(values["pg_port"])
	assert.Nil(t, err)

	covidDB := db.NewPostgresDB(values["pg_host"], port, values["pg_database"], values["pg_user"], values["pg_password"])
	assert.NotNil(t, covidDB)

	now := time.Now().UTC().Truncate(time.Second)

	newEntries := []db.CountryEntry{
		{
			Timestamp: now,
			Code:      "??",
			Name:      "???",
			Confirmed: 3,
			Deaths:    2,
			Recovered: 1,
		},
	}

	err = covidDB.Add(newEntries)
	assert.Nil(t, err)

	latest, err := covidDB.ListLatestByCountry()
	assert.Nil(t, err)
	latestTime, found := latest["???"]
	assert.True(t, found)
	assert.True(t, latestTime.Equal(now))

	allEntries, err := covidDB.List(time.Now())
	assert.Nil(t, err)

	found = false
	for _, entry := range allEntries {
		if entry.Timestamp.Equal(now) && entry.Name == "???" {
			assert.Equal(t, int64(3), entry.Confirmed)
			assert.Equal(t, int64(2), entry.Deaths)
			assert.Equal(t, int64(1), entry.Recovered)
			found = true
			break
		}
	}
	assert.True(t, found)
}
