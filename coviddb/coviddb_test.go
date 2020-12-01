package coviddb

import(
	"time"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestDBConnection (t *testing.T) {
	db := Create("192.168.0.11", 31000, "cicd", "cicd", "its4cicd")

	entries, err := db.List(time.Now())

	if err != nil {
		t.Error(err)
	} else if len(entries) == 0 {
		t.Error("No entries found")
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

	assert.Equal(t, 4,         len(allcases))
	assert.Equal(t, int64(1604188800000), allcases[CONFIRMED][0][TIMESTAMP])
	assert.Equal(t, int64(1604275200000), allcases[CONFIRMED][1][TIMESTAMP])
	assert.Equal(t, int64(1604448000000), allcases[CONFIRMED][2][TIMESTAMP])
	assert.Equal(t, int64(1),             allcases[CONFIRMED][0][VALUE])
	assert.Equal(t, int64(6),             allcases[CONFIRMED][1][VALUE])
	assert.Equal(t, int64(13),            allcases[CONFIRMED][2][VALUE])
	assert.Equal(t, int64(0),             allcases[RECOVERED][0][VALUE])
	assert.Equal(t, int64(1),             allcases[RECOVERED][1][VALUE])
	assert.Equal(t, int64(6),             allcases[RECOVERED][2][VALUE])
	assert.Equal(t, int64(0),             allcases[DEATHS][0][VALUE])
	assert.Equal(t, int64(0),             allcases[DEATHS][1][VALUE])
	assert.Equal(t, int64(1),             allcases[DEATHS][2][VALUE])

	deltas := GetTotalDeltas(allcases[CONFIRMED])

	assert.Equal(t, 3,        len(deltas))
	assert.Equal(t, int64(1604188800000), deltas[0][TIMESTAMP])
	assert.Equal(t, int64(1604275200000), deltas[1][TIMESTAMP])
	assert.Equal(t, int64(1604448000000), deltas[2][TIMESTAMP])
	assert.Equal(t, int64(1),             deltas[0][VALUE])
	assert.Equal(t, int64(5),             deltas[1][VALUE])
	assert.Equal(t, int64(7),             deltas[2][VALUE])
}


