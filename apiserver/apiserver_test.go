package apiserver

import(
	"time"
	"testing"
	"encoding/json"
	"github.com/stretchr/testify/assert"

	"covid19api/coviddb"
)

func parseDate(dateString string) (time.Time) {
        date, _ := time.Parse("2006-01-02", dateString)
        return date
}

func TestBuildSeries (t *testing.T) {
	entries := []coviddb.CountryEntry{
		coviddb.CountryEntry{
			Timestamp: parseDate("2020-11-01"),
			Code: "BE",
			Name: "Belgium",
			Confirmed: 1,
			Recovered: 0,
			Deaths: 0},
		coviddb.CountryEntry{
			Timestamp: parseDate("2020-11-02"),
			Code: "US",
			Name: "United States",
			Confirmed: 3,
			Recovered: 0,
			Deaths: 0},
		coviddb.CountryEntry{
			Timestamp: parseDate("2020-11-02"),
			Code: "BE",
			Name: "Belgium",
			Confirmed: 3,
			Recovered: 1,
			Deaths: 0},
		coviddb.CountryEntry{
			Timestamp: parseDate("2020-11-04"),
			Code: "US",
			Name: "United States",
			Confirmed: 10,
			Recovered: 5,
			Deaths: 1}}

	series := buildSeries(entries, []string{"confirmed", "confirmed-delta"})

	assert.Equal(t, 2,                 len(series))
	assert.Equal(t, "confirmed",       series[0].Target)
	assert.Equal(t, int64(1),          series[0].Datapoints[0][1])
	assert.Equal(t, int64(6),          series[0].Datapoints[1][1])
	assert.Equal(t, int64(13),         series[0].Datapoints[2][1])
	assert.Equal(t, "confirmed-delta", series[1].Target)
	assert.Equal(t, int64(1),          series[1].Datapoints[0][1])
	assert.Equal(t, int64(5),          series[1].Datapoints[1][1])
	assert.Equal(t, int64(7),          series[1].Datapoints[2][1])

	text, _ := json.Marshal(series)

	assert.Equal(t, "[{\"Target\":\"confirmed\",\"Datapoints\":[[1604188800000,1],[1604275200000,6],[1604448000000,13]]},{\"Target\":\"confirmed-delta\",\"Datapoints\":[[1604188800000,1],[1604275200000,5],[1604448000000,7]]}]", string(text))
}


