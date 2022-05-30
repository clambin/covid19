package store_test

import (
	"bou.ke/monkey"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/covid/store"
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	pg := configuration.LoadPGEnvironment()

	if !pg.IsValid() {
		fmt.Println("Could not find all CovidDB env variables. Skipping this test")
		return
	}

	var err error
	DB, err = db.NewWithConfiguration(pg)
	if err != nil {
		fmt.Printf("unable to connect to database: %s", err.Error())
		os.Exit(1)
	}

	covidStore = store.New(DB)

	m.Run()

	err = covidStore.(*store.PGCovidStore).RemoveDB()
	if err != nil {
		fmt.Printf("failed to clean up database: %s", err.Error())
		os.Exit(1)
	}
}

var (
	DB         *db.DB
	covidStore store.CovidStore
)

func TestDB(t *testing.T) {
	first := time.Date(2021, 12, 15, 0, 0, 0, 0, time.UTC)
	last := first.Add(24 * time.Hour)
	newEntries := []models.CountryEntry{
		{
			Timestamp: first,
			Code:      "??",
			Name:      "???",
			Confirmed: 3,
			Deaths:    2,
			Recovered: 1,
		},
		{
			Timestamp: last,
			Code:      "??",
			Name:      "???",
			Confirmed: 6,
			Deaths:    5,
			Recovered: 4,
		},
	}

	var (
		found     bool
		timestamp time.Time
	)

	entries, err := covidStore.GetAll()
	require.NoError(t, err)
	assert.Len(t, entries, 0)

	_, found, err = covidStore.GetFirstEntry()
	require.NoError(t, err)
	assert.False(t, found)

	err = covidStore.Add(newEntries)
	require.NoError(t, err)

	timestamp, found, err = covidStore.GetFirstEntry()
	require.NoError(t, err)
	require.True(t, found)
	assert.True(t, timestamp.Equal(first))

	entries, err = covidStore.GetAll()
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.True(t, entries[0].Timestamp.Equal(first))
	assert.Equal(t, int64(3), entries[0].Confirmed)
	assert.Equal(t, int64(2), entries[0].Deaths)
	assert.Equal(t, int64(1), entries[0].Recovered)
	assert.True(t, entries[1].Timestamp.Equal(last))
	assert.Equal(t, int64(6), entries[1].Confirmed)
	assert.Equal(t, int64(5), entries[1].Deaths)
	assert.Equal(t, int64(4), entries[1].Recovered)

	entries, err = covidStore.GetAllForRange(first, first)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.True(t, entries[0].Timestamp.Equal(first))

	entries, err = covidStore.GetAllForCountryName("???")
	require.NoError(t, err)
	assert.Len(t, entries, 2)

	var countryNames []string
	countryNames, err = covidStore.GetAllCountryNames()
	require.NoError(t, err)
	require.Len(t, countryNames, 1)
	assert.Equal(t, "???", countryNames[0])

	var latest map[string]models.CountryEntry
	latest, err = covidStore.GetLatestForCountries([]string{"???"})
	require.NoError(t, err)
	entry, found := latest["???"]
	require.True(t, found)
	assert.True(t, entry.Timestamp.Equal(last))
	assert.Equal(t, int64(6), entry.Confirmed)
	assert.Equal(t, int64(5), entry.Deaths)
	assert.Equal(t, int64(4), entry.Recovered)

	latest, err = covidStore.GetLatestForCountriesByTime([]string{"???"}, first)
	require.NoError(t, err)
	entry, found = latest["???"]
	require.True(t, found)
	assert.True(t, entry.Timestamp.Equal(first))
	assert.Equal(t, int64(3), entry.Confirmed)
	assert.Equal(t, int64(2), entry.Deaths)
	assert.Equal(t, int64(1), entry.Recovered)

	updates, err := covidStore.CountEntriesByTime(first, last)
	require.NoError(t, err)
	assert.Len(t, updates, 2)
	for ts, value := range updates {
		assert.True(t, ts.Equal(first) || ts.Equal(last))
		assert.Equal(t, 1, value)
	}
}

func TestNew_Failure(t *testing.T) {
	db2, _ := db.New("localhost", 0, "covid", "covid", "password")

	fakeExit := func(int) {
		panic("os.Exit called")
	}
	patch := monkey.Patch(os.Exit, fakeExit)
	defer patch.Unpatch()

	assert.Panics(t, func() {
		_ = store.New(db2)
	})
}
