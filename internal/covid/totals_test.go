package covid_test

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"covid19/internal/covid"
	"covid19/internal/coviddb"
)

var (
	testData = []coviddb.CountryEntry{
		{
			Timestamp: parseDate("2020-11-01"),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0},
		{
			Timestamp: parseDate("2020-11-02"),
			Code:      "US",
			Name:      "United States",
			Confirmed: 3,
			Recovered: 0,
			Deaths:    0},
		{
			Timestamp: parseDate("2020-11-02"),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 3,
			Recovered: 1,
			Deaths:    0},
		{
			Timestamp: parseDate("2020-11-04"),
			Code:      "US",
			Name:      "United States",
			Confirmed: 10,
			Recovered: 5,
			Deaths:    1}}
)

func TestTotalAndDelta(t *testing.T) {
	entries := testData
	allcases := covid.GetTotalCases(entries)

	assert.Equal(t, 4, len(allcases))
	assert.Equal(t, int64(1604188800000), allcases[covid.CONFIRMED][0][covid.TIMESTAMP])
	assert.Equal(t, int64(1604275200000), allcases[covid.CONFIRMED][1][covid.TIMESTAMP])
	assert.Equal(t, int64(1604448000000), allcases[covid.CONFIRMED][2][covid.TIMESTAMP])
	assert.Equal(t, int64(1), allcases[covid.CONFIRMED][0][covid.VALUE])
	assert.Equal(t, int64(6), allcases[covid.CONFIRMED][1][covid.VALUE])
	assert.Equal(t, int64(13), allcases[covid.CONFIRMED][2][covid.VALUE])
	assert.Equal(t, int64(0), allcases[covid.RECOVERED][0][covid.VALUE])
	assert.Equal(t, int64(1), allcases[covid.RECOVERED][1][covid.VALUE])
	assert.Equal(t, int64(6), allcases[covid.RECOVERED][2][covid.VALUE])
	assert.Equal(t, int64(0), allcases[covid.DEATHS][0][covid.VALUE])
	assert.Equal(t, int64(0), allcases[covid.DEATHS][1][covid.VALUE])
	assert.Equal(t, int64(1), allcases[covid.DEATHS][2][covid.VALUE])

	deltas := covid.GetTotalDeltas(allcases[covid.CONFIRMED])

	assert.Equal(t, 3, len(deltas))
	assert.Equal(t, int64(1604188800000), deltas[0][covid.TIMESTAMP])
	assert.Equal(t, int64(1604275200000), deltas[1][covid.TIMESTAMP])
	assert.Equal(t, int64(1604448000000), deltas[2][covid.TIMESTAMP])
	assert.Equal(t, int64(1), deltas[0][covid.VALUE])
	assert.Equal(t, int64(5), deltas[1][covid.VALUE])
	assert.Equal(t, int64(7), deltas[2][covid.VALUE])
}
