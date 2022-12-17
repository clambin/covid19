package simplejsonserver_test

import (
	"github.com/clambin/covid19/configuration"
	mockCovidStore "github.com/clambin/covid19/db/mocks"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	cfg := configuration.Configuration{}
	covidDB := mockCovidStore.NewCovidStore(t)
	popDB := mockCovidStore.NewPopulationStore(t)
	s, err := simplejsonserver.New(&cfg, covidDB, popDB)
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodPost, "/search", nil)
	resp := httptest.NewRecorder()

	s.Search(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `["country-confirmed","country-confirmed-population","country-deaths","country-deaths-population","country-deaths-vs-confirmed","cumulative","evolution","incremental","updates"]`, string(body))

	covidDB.On("GetLatestForCountriesByTime", mock.AnythingOfType("time.Time")).Return(map[string]models.CountryEntry{
		"A": {Timestamp: time.Date(2022, 1, 19, 0, 0, 0, 0, time.UTC), Code: "A", Confirmed: 4, Deaths: 1},
		"B": {Timestamp: time.Date(2022, 1, 19, 0, 0, 0, 0, time.UTC), Code: "B", Confirmed: 10, Deaths: 5},
	}, nil)
	covidDB.On("GetTotalsPerDay").Return([]models.CountryEntry{
		{
			Timestamp: time.Date(2022, 1, 18, 0, 0, 0, 0, time.UTC),
		},
		{
			Timestamp: time.Date(2022, 1, 19, 0, 0, 0, 0, time.UTC),
			Confirmed: 14,
			Deaths:    6,
		},
	}, nil)
	covidDB.
		On("GetAllForRange", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return([]models.CountryEntry{
			{
				Timestamp: time.Date(2022, time.January, 18, 0, 0, 0, 0, time.UTC),
				Code:      "A",
				Name:      "A",
			},
			{
				Timestamp: time.Date(2022, time.January, 19, 0, 0, 0, 0, time.UTC),
				Code:      "A",
				Name:      "A",
				Confirmed: 4,
				Deaths:    1,
			},
			{
				Timestamp: time.Date(2022, time.January, 18, 0, 0, 0, 0, time.UTC),
				Code:      "B",
				Name:      "B",
			},
			{
				Timestamp: time.Date(2022, time.January, 19, 0, 0, 0, 0, time.UTC),
				Code:      "B",
				Name:      "B",
				Confirmed: 10,
				Deaths:    5,
			},
		}, nil)
	covidDB.
		On("GetAllForCountryName", "A").
		Return([]models.CountryEntry{
			{
				Timestamp: time.Date(2022, time.January, 18, 0, 0, 0, 0, time.UTC),
				Code:      "A",
				Name:      "A",
			},
			{
				Timestamp: time.Date(2022, time.January, 19, 0, 0, 0, 0, time.UTC),
				Code:      "A",
				Name:      "A",
				Confirmed: 4,
				Deaths:    1,
			}}, nil)
	covidDB.
		On("CountEntriesByTime", time.Date(2022, time.January, 20, 0, 0, 0, 0, time.UTC), time.Date(2022, time.January, 20, 0, 0, 0, 0, time.UTC)).
		Return([]struct {
			Timestamp time.Time
			Count     int
		}{
			{Timestamp: time.Date(2022, time.January, 20, 0, 0, 0, 0, time.UTC), Count: 2},
		},
			nil)

	popDB.On("List").Return(map[string]int64{
		"A": 10,
		"B": 100,
	}, nil)

	var testCases = []struct {
		input  string
		output string
		fail   bool
	}{
		{
			input: `{"targets": [{"target": "country-confirmed","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"A","type":"number"},{"text":"B","type":"number"}],"rows":[["2022-01-19T00:00:00Z",4,10]]}]
`,
		},
		{
			input: `{"targets": [{"target": "country-deaths","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"A","type":"number"},{"text":"B","type":"number"}],"rows":[["2022-01-19T00:00:00Z",1,5]]}]
`,
		},
		{
			input: `{"targets": [{"target": "country-confirmed-population","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"country","type":"string"},{"text":"confirmed","type":"number"}],"rows":[["2022-01-19T00:00:00Z","A",0.4],["2022-01-19T00:00:00Z","B",0.1]]}]
`,
		},
		{
			input: `{"targets": [{"target": "country-deaths-population","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"country","type":"string"},{"text":"deaths","type":"number"}],"rows":[["2022-01-19T00:00:00Z","A",0.1],["2022-01-19T00:00:00Z","B",0.05]]}]
`,
		},
		{
			input: `{"targets": [{"target": "country-deaths-vs-confirmed","type": "table"}],"range": {"to": "2022-01-17T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"country","type":"string"},{"text":"ratio","type":"number"}],"rows":[["2022-01-19T00:00:00Z","A",0.25],["2022-01-19T00:00:00Z","B",0.5]]}]
`,
		},
		{
			input: `{"targets": [{"target": "incremental","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"confirmed","type":"number"},{"text":"deaths","type":"number"}],"rows":[["2022-01-18T00:00:00Z",0,0],["2022-01-19T00:00:00Z",14,6]]}]
`,
		},
		{
			input: `{"targets": [{"target": "incremental","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"},"adhocFilters": [{"key": "Country Name","operator": "=","value": "A"}]}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"confirmed","type":"number"},{"text":"deaths","type":"number"}],"rows":[["2022-01-18T00:00:00Z",0,0],["2022-01-19T00:00:00Z",4,1]]}]
`,
		},
		{
			input: `{"targets": [{"target": "cumulative","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"confirmed","type":"number"},{"text":"deaths","type":"number"}],"rows":[["2022-01-18T00:00:00Z",0,0],["2022-01-19T00:00:00Z",14,6]]}]
`,
		},
		{
			input: `{"targets": [{"target": "cumulative","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"},"adhocFilters": [{"key": "Country Name","operator": "=","value": "A"}]}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"confirmed","type":"number"},{"text":"deaths","type":"number"}],"rows":[["2022-01-18T00:00:00Z",0,0],["2022-01-19T00:00:00Z",4,1]]}]
`,
		},
		{
			input: `{"targets": [{"target": "evolution","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"country","type":"string"},{"text":"increase","type":"number"}],"rows":[["2022-01-20T00:00:00Z","A",4],["2022-01-20T00:00:00Z","B",10]]}]
`,
		},
		{
			input: `{"targets": [{"target": "updates","type": "table"}],"range": {"from": "2022-01-20T00:00:00Z","to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"updates","type":"number"}],"rows":[["2022-01-20T00:00:00Z",2]]}]
`,
		},
	}

	for index, tt := range testCases {
		req, _ = http.NewRequest(http.MethodPost, "/query", strings.NewReader(tt.input))
		resp = httptest.NewRecorder()
		s.Query(resp, req)

		if tt.fail {
			require.NotEqual(t, http.StatusOK, resp.Code)
			continue
		}

		require.Equal(t, http.StatusOK, resp.Code)
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, index)
		assert.Equal(t, tt.output, string(body), tt.input)
	}
}
