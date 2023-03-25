package simplejsonserver_test

import (
	"github.com/clambin/covid19/internal/testtools/db/covid"
	"github.com/clambin/covid19/internal/testtools/db/population"
	"github.com/clambin/covid19/models"
	"github.com/clambin/covid19/simplejsonserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var _ simplejsonserver.CovidGetter = &covid.FakeStore{}

func TestServer(t *testing.T) {
	covidDB := covid.FakeStore{Records: []models.CountryEntry{
		{Timestamp: time.Date(2022, time.January, 18, 0, 0, 0, 0, time.UTC), Code: "A", Name: "A"},
		{Timestamp: time.Date(2022, time.January, 19, 0, 0, 0, 0, time.UTC), Code: "A", Name: "A", Confirmed: 4, Deaths: 1},
		{Timestamp: time.Date(2022, time.January, 18, 0, 0, 0, 0, time.UTC), Code: "B", Name: "B"},
		{Timestamp: time.Date(2022, time.January, 19, 0, 0, 0, 0, time.UTC), Code: "B", Name: "B", Confirmed: 10, Deaths: 5},
	}}
	popDB := population.FakeStore{Content: map[string]int64{
		"A": 10,
		"B": 100,
	}}
	s := simplejsonserver.New(&covidDB, &popDB)

	req, _ := http.NewRequest(http.MethodPost, "/search", nil)
	resp := httptest.NewRecorder()

	s.Search(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `["country-confirmed","country-confirmed-population","country-deaths","country-deaths-population","country-deaths-vs-confirmed","cumulative","evolution","incremental","updates"]`, string(body))

	var testCases = []struct {
		name   string
		input  string
		output string
		fail   bool
	}{
		{
			name:  "country-confirmed",
			input: `{"targets": [{"target": "country-confirmed","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"A","type":"number"},{"text":"B","type":"number"}],"rows":[["2022-01-19T00:00:00Z",4,10]]}]
`,
		},
		{
			name:  "country-deaths",
			input: `{"targets": [{"target": "country-deaths","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"A","type":"number"},{"text":"B","type":"number"}],"rows":[["2022-01-19T00:00:00Z",1,5]]}]
`,
		},
		{
			name:  "country-confirmed-population",
			input: `{"targets": [{"target": "country-confirmed-population","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"country","type":"string"},{"text":"confirmed","type":"number"}],"rows":[["2022-01-19T00:00:00Z","A",0.4],["2022-01-19T00:00:00Z","B",0.1]]}]
`,
		},
		{
			name:  "country-deaths-population",
			input: `{"targets": [{"target": "country-deaths-population","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"country","type":"string"},{"text":"deaths","type":"number"}],"rows":[["2022-01-19T00:00:00Z","A",0.1],["2022-01-19T00:00:00Z","B",0.05]]}]
`,
		},
		{
			name:  "country-deaths-vs-confirmed",
			input: `{"targets": [{"target": "country-deaths-vs-confirmed","type": "table"}],"range": {"to": "2022-01-19T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"country","type":"string"},{"text":"ratio","type":"number"}],"rows":[["2022-01-19T00:00:00Z","A",0.25],["2022-01-19T00:00:00Z","B",0.5]]}]
`,
		},
		{
			name:  "incremental",
			input: `{"targets": [{"target": "incremental","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"confirmed","type":"number"},{"text":"deaths","type":"number"}],"rows":[["2022-01-18T00:00:00Z",0,0],["2022-01-19T00:00:00Z",14,6]]}]
`,
		},
		{
			name:  "incremental (filtered)",
			input: `{"targets": [{"target": "incremental","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"},"adhocFilters": [{"key": "Country Name","operator": "=","value": "A"}]}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"confirmed","type":"number"},{"text":"deaths","type":"number"}],"rows":[["2022-01-18T00:00:00Z",0,0],["2022-01-19T00:00:00Z",4,1]]}]
`,
		},
		{
			name:  "cumulative",
			input: `{"targets": [{"target": "cumulative","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"confirmed","type":"number"},{"text":"deaths","type":"number"}],"rows":[["2022-01-18T00:00:00Z",0,0],["2022-01-19T00:00:00Z",14,6]]}]
`,
		},
		{
			name:  "cumulative (filtered)",
			input: `{"targets": [{"target": "cumulative","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"},"adhocFilters": [{"key": "Country Name","operator": "=","value": "A"}]}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"confirmed","type":"number"},{"text":"deaths","type":"number"}],"rows":[["2022-01-18T00:00:00Z",0,0],["2022-01-19T00:00:00Z",4,1]]}]
`,
		},
		{
			name:  "evolution",
			input: `{"targets": [{"target": "evolution","type": "table"}],"range": {"to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"country","type":"string"},{"text":"increase","type":"number"}],"rows":[["2022-01-20T00:00:00Z","A",4],["2022-01-20T00:00:00Z","B",10]]}]
`,
		},
		{
			name:  "updates",
			input: `{"targets": [{"target": "updates","type": "table"}],"range": {"from": "2022-01-19T00:00:00Z","to": "2022-01-19T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"updates","type":"number"}],"rows":[["2022-01-19T00:00:00Z",2]]}]
`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req, _ = http.NewRequest(http.MethodPost, "/query", strings.NewReader(tt.input))
			resp = httptest.NewRecorder()
			s.Query(resp, req)

			if tt.fail {
				require.NotEqual(t, http.StatusOK, resp.Code)
				return
			}

			require.Equal(t, http.StatusOK, resp.Code)
			body, err = io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.output, string(body))
		})
	}
}
