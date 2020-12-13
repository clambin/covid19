package covidhandler_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"covid19/internal/covid/db"
	"covid19/internal/covidhandler"
)

var (
	testData = []db.CountryEntry{
		{
			Timestamp: time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 1,
			Recovered: 0,
			Deaths:    0},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Code:      "US",
			Name:      "United States",
			Confirmed: 3,
			Recovered: 0,
			Deaths:    0},
		{
			Timestamp: time.Date(2020, time.November, 2, 0, 0, 0, 0, time.UTC),
			Code:      "BE",
			Name:      "Belgium",
			Confirmed: 3,
			Recovered: 1,
			Deaths:    0},
		{
			Timestamp: time.Date(2020, time.November, 4, 0, 0, 0, 0, time.UTC),
			Code:      "US",
			Name:      "United States",
			Confirmed: 10,
			Recovered: 5,
			Deaths:    1}}
)

func TestTotalAndDelta(t *testing.T) {
	entries := testData
	allcases := covidhandler.GetTotalCases(entries)

	assert.Equal(t, 4, len(allcases))
	assert.Equal(t, int64(1604188800000), allcases[covidhandler.CONFIRMED][0][covidhandler.TIMESTAMP])
	assert.Equal(t, int64(1604275200000), allcases[covidhandler.CONFIRMED][1][covidhandler.TIMESTAMP])
	assert.Equal(t, int64(1604448000000), allcases[covidhandler.CONFIRMED][2][covidhandler.TIMESTAMP])
	assert.Equal(t, int64(1), allcases[covidhandler.CONFIRMED][0][covidhandler.VALUE])
	assert.Equal(t, int64(6), allcases[covidhandler.CONFIRMED][1][covidhandler.VALUE])
	assert.Equal(t, int64(13), allcases[covidhandler.CONFIRMED][2][covidhandler.VALUE])
	assert.Equal(t, int64(0), allcases[covidhandler.RECOVERED][0][covidhandler.VALUE])
	assert.Equal(t, int64(1), allcases[covidhandler.RECOVERED][1][covidhandler.VALUE])
	assert.Equal(t, int64(6), allcases[covidhandler.RECOVERED][2][covidhandler.VALUE])
	assert.Equal(t, int64(0), allcases[covidhandler.DEATHS][0][covidhandler.VALUE])
	assert.Equal(t, int64(0), allcases[covidhandler.DEATHS][1][covidhandler.VALUE])
	assert.Equal(t, int64(1), allcases[covidhandler.DEATHS][2][covidhandler.VALUE])

	deltas := covidhandler.GetTotalDeltas(allcases[covidhandler.CONFIRMED])

	assert.Equal(t, 3, len(deltas))
	assert.Equal(t, int64(1604188800000), deltas[0][covidhandler.TIMESTAMP])
	assert.Equal(t, int64(1604275200000), deltas[1][covidhandler.TIMESTAMP])
	assert.Equal(t, int64(1604448000000), deltas[2][covidhandler.TIMESTAMP])
	assert.Equal(t, int64(1), deltas[0][covidhandler.VALUE])
	assert.Equal(t, int64(5), deltas[1][covidhandler.VALUE])
	assert.Equal(t, int64(7), deltas[2][covidhandler.VALUE])
}
