package coviddb_test

import (
	"github.com/clambin/covid19/coviddb"
	db2 "github.com/clambin/covid19/db"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
	"time"
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
	assert.NoError(t, err)

	var DB *db2.DB
	DB, err = db2.New(values["pg_host"], port, values["pg_database"], values["pg_user"], values["pg_password"])
	assert.NoError(t, err)

	var covidDB *coviddb.PostgresDB
	covidDB, err = coviddb.New(DB)
	assert.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Second)

	newEntries := []coviddb.CountryEntry{
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
	assert.NoError(t, err)

	var latest map[string]time.Time
	latest, err = covidDB.ListLatestByCountry()
	assert.NoError(t, err)
	latestTime, found := latest["???"]
	assert.True(t, found)
	assert.True(t, latestTime.Equal(now))

	var allEntries []coviddb.CountryEntry
	allEntries, err = covidDB.List(time.Now())
	assert.NoError(t, err)

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

	var first time.Time
	first, found, err = covidDB.GetFirstEntry()
	assert.NoError(t, err)
	assert.True(t, found)
	assert.True(t, first.Equal(now))

	var entry *coviddb.CountryEntry
	entry, found, err = covidDB.GetLastBeforeDate("???", now.AddDate(1, 0, 0))

	assert.NoError(t, err)
	assert.True(t, found)
	if assert.NotNil(t, entry) {
		assert.Equal(t, "???", entry.Name)
		assert.Equal(t, int64(3), entry.Confirmed)
		assert.Equal(t, int64(2), entry.Deaths)
		assert.Equal(t, int64(1), entry.Recovered)
	}

	entry, found, err = covidDB.GetLastBeforeDate("???", now.AddDate(-1, 0, 0))
	assert.NoError(t, err)
	assert.False(t, found)

	var codes []string
	codes, err = covidDB.GetAllCountryCodes()
	assert.NoError(t, err)
	if assert.Len(t, codes, 1) {
		assert.Equal(t, "??", codes[0])
	}

	err = covidDB.RemoveDB()
	assert.NoError(t, err)

	_, err = covidDB.ListLatestByCountry()
	assert.Error(t, err)
}
