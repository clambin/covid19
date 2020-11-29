package coviddb

import(
	"time"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestDBConnection (t *testing.T) {
	pgdb, err := Connect("192.168.0.11", 31000, "cicd", "cicd", "its4cicd")
	if err == nil {
		entries, _ := pgdb.List()
		pgdb.Close()
		if len(entries) == 0 {
			t.Error("No entries found")
		}
	} else {
		t.Error(err)
	}
}

func parseDate(dateString string) (time.Time) {
	date, _ := time.Parse("2006-01-02", dateString)
	return date
}

func TestTotalAndDelta (t *testing.T) {
	entries := []CountryEntry{
		CountryEntry{
			Timestamp: parseDate("2020-11-01"),
			Code: "BE",
			Name: "Belgium",
			Confirmed: 1,
			Recovered: 0,
			Deaths: 0},
		CountryEntry{
			Timestamp: parseDate("2020-11-02"),
			Code: "US",
			Name: "United States",
			Confirmed: 3,
			Recovered: 0,
			Deaths: 0},
		CountryEntry{
			Timestamp: parseDate("2020-11-02"),
			Code: "BE",
			Name: "Belgium",
			Confirmed: 3,
			Recovered: 1,
			Deaths: 0},
		CountryEntry{
			Timestamp: parseDate("2020-11-04"),
			Code: "US",
			Name: "United States",
			Confirmed: 10,
			Recovered: 5,
			Deaths: 1}}

	allcases := GetTotalCases(entries)

	assert.Equal(t, 3,         len(allcases))
	assert.Equal(t, parseDate("2020-11-01"), allcases[0].Timestamp)
	assert.Equal(t, parseDate("2020-11-02"), allcases[1].Timestamp)
	assert.Equal(t, parseDate("2020-11-04"), allcases[2].Timestamp)
	assert.Equal(t, int64(1),  allcases[0].Confirmed)
	assert.Equal(t, int64(6),  allcases[1].Confirmed)
	assert.Equal(t, int64(13), allcases[2].Confirmed)
	assert.Equal(t, int64(0),  allcases[0].Recovered)
	assert.Equal(t, int64(1),  allcases[1].Recovered)
	assert.Equal(t, int64(6),  allcases[2].Recovered)
	assert.Equal(t, int64(0),  allcases[0].Deaths)
	assert.Equal(t, int64(0),  allcases[1].Deaths)
	assert.Equal(t, int64(1),  allcases[2].Deaths)

	alldeltas := GetTotalDeltas(allcases)

	assert.Equal(t, 3,        len(alldeltas))
	assert.Equal(t, parseDate("2020-11-01"), alldeltas[0].Timestamp)
	assert.Equal(t, parseDate("2020-11-02"), alldeltas[1].Timestamp)
	assert.Equal(t, parseDate("2020-11-04"), alldeltas[2].Timestamp)
	assert.Equal(t, int64(1), alldeltas[0].Confirmed)
	assert.Equal(t, int64(5), alldeltas[1].Confirmed)
	assert.Equal(t, int64(7), alldeltas[2].Confirmed)
	assert.Equal(t, int64(0), alldeltas[0].Recovered)
	assert.Equal(t, int64(1), alldeltas[1].Recovered)
	assert.Equal(t, int64(5), alldeltas[2].Recovered)
	assert.Equal(t, int64(0), alldeltas[0].Deaths)
	assert.Equal(t, int64(0), alldeltas[1].Deaths)
	assert.Equal(t, int64(1), alldeltas[2].Deaths)
}


