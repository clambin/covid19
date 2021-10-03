package store_test

import (
	"bou.ke/monkey"
	"github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	var DB *db.DB
	DB, err = db.New(values["pg_host"], port, values["pg_database"], values["pg_user"], values["pg_password"])
	require.NoError(t, err)

	var covidStore store.CovidStore
	covidStore = store.New(DB)
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Second)

	newEntries := []*models.CountryEntry{
		{
			Timestamp: now,
			Code:      "??",
			Name:      "???",
			Confirmed: 3,
			Deaths:    2,
			Recovered: 1,
		},
	}

	err = covidStore.Add(newEntries)
	require.NoError(t, err)

	var found bool
	// var timestamp time.Time
	_, found, err = covidStore.GetFirstEntry()
	require.NoError(t, err)
	require.True(t, found)
	// assert.Equal(t, now, timestamp)

	var entries []*models.CountryEntry
	entries, err = covidStore.GetAll()
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.True(t, entries[0].Timestamp.Equal(now))
	assert.Equal(t, int64(3), entries[0].Confirmed)
	assert.Equal(t, int64(2), entries[0].Deaths)
	assert.Equal(t, int64(1), entries[0].Recovered)

	var latest map[string]*models.CountryEntry
	latest, err = covidStore.GetLatestForCountries([]string{"???"})
	require.NoError(t, err)
	entry, found := latest["???"]
	require.True(t, found)
	assert.True(t, entry.Timestamp.Equal(now))
	assert.Equal(t, int64(3), entry.Confirmed)
	assert.Equal(t, int64(2), entry.Deaths)
	assert.Equal(t, int64(1), entry.Recovered)

	err = covidStore.(*store.PGCovidStore).RemoveDB()
	assert.NoError(t, err)

	// reinitialize
	covidStore = store.New(DB)

	entries, err = covidStore.GetAll()
	require.NoError(t, err)
	assert.Len(t, entries, 0)

	_, found, err = covidStore.GetFirstEntry()
	require.NoError(t, err)
	assert.False(t, found)
}

func TestNew_Failure(t *testing.T) {
	DB, _ := db.New("localhost", 0, "covid", "covid", "password")

	fakeExit := func(int) {
		panic("os.Exit called")
	}
	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	assert.Panics(t, func() {
		_ = store.New(DB)
	})
}
