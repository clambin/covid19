package simplejsonserver_test

import (
	"context"
	"errors"
	"fmt"
	mockCovidStore "github.com/clambin/covid19/covid/store/mocks"
	"github.com/clambin/covid19/models"
	mockPopulationStore "github.com/clambin/covid19/population/store/mocks"
	"github.com/clambin/covid19/simplejsonserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestServer_Query(t *testing.T) {
	covidDB := &mockCovidStore.CovidStore{}
	popDB := &mockPopulationStore.PopulationStore{}
	s := simplejsonserver.MakeServer(covidDB, popDB)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := s.Run(8080)
		require.True(t, errors.Is(err, http.ErrServerClosed))
		wg.Done()
	}()

	require.Eventually(t, func() bool {
		_, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", 8080))
		return err == nil

	}, time.Second, 10*time.Millisecond)

	covidDB.On("GetAllCountryNames").Return([]string{"A", "B"}, nil)
	covidDB.On("GetLatestForCountriesByTime", []string{"A", "B"}, mock.AnythingOfType("time.Time")).Return(map[string]models.CountryEntry{
		"A": {Timestamp: time.Date(2022, 1, 19, 0, 0, 0, 0, time.UTC), Code: "A", Confirmed: 4, Deaths: 1},
		"B": {Timestamp: time.Date(2022, 1, 19, 0, 0, 0, 0, time.UTC), Code: "B", Confirmed: 10, Deaths: 5},
	}, nil)
	covidDB.On("GetAll").Return([]models.CountryEntry{
		{
			Timestamp: time.Date(2022, 1, 18, 0, 0, 0, 0, time.UTC),
			Code:      "A",
			Name:      "A",
		},
		{
			Timestamp: time.Date(2022, 1, 19, 0, 0, 0, 0, time.UTC),
			Code:      "A",
			Name:      "A",
			Confirmed: 4,
			Deaths:    1,
		},
		{
			Timestamp: time.Date(2022, 1, 18, 0, 0, 0, 0, time.UTC),
			Code:      "B",
			Name:      "B",
		},
		{
			Timestamp: time.Date(2022, 1, 19, 0, 0, 0, 0, time.UTC),
			Code:      "B",
			Name:      "B",
			Confirmed: 10,
			Deaths:    5,
		},
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
		On("GetAllForRange",
			time.Date(2022, time.January, 20, 0, 0, 0, 0, time.UTC),
			time.Date(2022, time.January, 20, 0, 0, 0, 0, time.UTC),
		).
		Return([]models.CountryEntry{
			{
				Timestamp: time.Date(2022, time.January, 20, 0, 0, 0, 0, time.UTC),
				Code:      "A",
				Name:      "A",
				Confirmed: 4,
				Deaths:    1,
			},
			{
				Timestamp: time.Date(2022, time.January, 20, 0, 0, 0, 0, time.UTC),
				Code:      "B",
				Name:      "B",
				Confirmed: 10,
				Deaths:    5,
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
		Return(map[time.Time]int{time.Date(2022, time.January, 20, 0, 0, 0, 0, time.UTC): 2}, nil)

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
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"country","type":"string"},{"text":"increase","type":"number"}],"rows":[["2022-01-20T00:00:00Z","A",2],["2022-01-20T00:00:00Z","B",5]]}]
`,
		},
		{
			input: `{"targets": [{"target": "updates","type": "table"}],"range": {"from": "2022-01-20T00:00:00Z","to": "2022-01-20T00:00:00Z"}}`,
			output: `[{"type":"table","columns":[{"text":"timestamp","type":"time"},{"text":"updates","type":"number"}],"rows":[["2022-01-20T00:00:00Z",2]]}]
`,
		},
	}

	for index, testCase := range testCases {
		resp, err := http.Post(
			fmt.Sprintf("http://127.0.0.1:%d/query", 8080),
			"application/json",
			strings.NewReader(testCase.input),
		)
		if testCase.fail {
			require.Error(t, err, index)
			continue
		}
		require.NoError(t, err, index)
		var body []byte
		body, err = io.ReadAll(resp.Body)
		require.NoError(t, err, index)
		assert.Equal(t, testCase.output, string(body), testCase.input)
	}

	err := s.Shutdown(context.Background(), 15*time.Second)
	require.NoError(t, err)
	wg.Wait()
}

func TestServer_Search(t *testing.T) {
	s := simplejsonserver.MakeServer(nil, nil)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := s.Run(8080)
		require.True(t, errors.Is(err, http.ErrServerClosed))
		wg.Done()
	}()

	require.Eventually(t, func() bool {
		_, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", 8080))
		return err == nil

	}, time.Second, 10*time.Millisecond)

	resp, err := http.Post(
		fmt.Sprintf("http://127.0.0.1:%d/search", 8080),
		"application/json",
		nil,
	)
	require.NoError(t, err)
	var body []byte
	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `["country-confirmed","country-confirmed-population","country-deaths","country-deaths-population","country-deaths-vs-confirmed","cumulative","evolution","incremental","updates"]`, string(body))
}
