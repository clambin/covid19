package covidhandler

import (
	"errors"
	"fmt"
	"github.com/clambin/covid19/internal/covidcache"
	"github.com/clambin/grafana-json"
	log "github.com/sirupsen/logrus"
	"strings"
)

// Handler implements business logic for APIServer
type Handler struct {
	cache *covidcache.Cache
}

// Create a CovidAPIHandler object
func Create(cache *covidcache.Cache) (*Handler, error) {
	if cache == nil {
		return nil, errors.New("no database specified")
	}
	return &Handler{cache: cache}, nil
}

var (
	Targets = []string{
		"active",
		"active-delta",
		"confirmed",
		"confirmed-delta",
		"death",
		"death-delta",
		"recovered",
		"recovered-delta",
		"daily",
		"cumulative",
	}
)

// Search returns all supported targets
func (handler *Handler) Search() []string {
	return Targets
}

// Query the DB and return the requested targets
func (handler *Handler) Query(target string, request *grafana_json.QueryRequest) (response *grafana_json.QueryResponse, err error) {
	responseChannel := make(chan covidcache.Response)
	defer close(responseChannel)

	deltas := false
	subTarget := target
	if strings.HasSuffix(target, "-delta") {
		deltas = true
		subTarget = strings.TrimSuffix(target, "-delta")
	}

	handler.cache.Request <- covidcache.Request{
		Response: responseChannel,
		End:      request.Range.To,
		Delta:    deltas,
	}

	resp := <-responseChannel

	response = new(grafana_json.QueryResponse)
	response.Target = target
	response.DataPoints = make([]grafana_json.QueryResponseDataPoint, len(resp.Series))

loop:
	for index, entry := range resp.Series {
		var value int64
		switch subTarget {
		case "confirmed":
			value = entry.Confirmed
		case "recovered":
			value = entry.Recovered
		case "death":
			value = entry.Deaths
		case "active":
			value = entry.Active
		default:
			log.Warningf("dropping unsupported target: %s", target)
			break loop
		}

		response.DataPoints[index] = grafana_json.QueryResponseDataPoint{
			Timestamp: entry.Timestamp,
			Value:     value,
		}
	}
	return
}

func (handler *Handler) QueryTable(target string, req *grafana_json.QueryRequest) (response *grafana_json.QueryTableResponse, err error) {
	switch target {
	case "daily":
		response = handler.buildDaily(req)
	case "cumulative":
		response = handler.buildCumulative(req)
	default:
		err = fmt.Errorf("%s does not implement table query", target)
	}
	return
}

func (handler *Handler) buildDaily(request *grafana_json.QueryRequest) (response *grafana_json.QueryTableResponse) {
	responseChannel := make(chan covidcache.Response)
	defer close(responseChannel)

	handler.cache.Request <- covidcache.Request{
		Response: responseChannel,
		End:      request.Range.To,
		Delta:    true,
	}

	resp := <-responseChannel

	dataLen := len(resp.Series)
	response = new(grafana_json.QueryTableResponse)
	timestamps := make(grafana_json.QueryTableResponseTimeColumn, 0, dataLen)
	confirmed := make(grafana_json.QueryTableResponseNumberColumn, 0, dataLen)
	recovered := make(grafana_json.QueryTableResponseNumberColumn, 0, dataLen)
	deaths := make(grafana_json.QueryTableResponseNumberColumn, 0, dataLen)

	for _, entry := range resp.Series {
		timestamps = append(timestamps, entry.Timestamp)
		confirmed = append(confirmed, float64(entry.Confirmed))
		recovered = append(recovered, float64(entry.Recovered))
		deaths = append(deaths, float64(entry.Deaths))
	}

	response = new(grafana_json.QueryTableResponse)
	response.Columns = []grafana_json.QueryTableResponseColumn{
		{Text: "timestamp", Data: timestamps},
		{Text: "confirmed", Data: confirmed},
		{Text: "recovered", Data: recovered},
		{Text: "deaths", Data: deaths},
	}
	return
}

func (handler *Handler) buildCumulative(request *grafana_json.QueryRequest) (response *grafana_json.QueryTableResponse) {
	responseChannel := make(chan covidcache.Response)
	defer close(responseChannel)

	handler.cache.Request <- covidcache.Request{
		Response: responseChannel,
		End:      request.Range.To,
		Delta:    false,
	}

	resp := <-responseChannel

	dataLen := len(resp.Series)
	response = new(grafana_json.QueryTableResponse)
	timestamps := make(grafana_json.QueryTableResponseTimeColumn, 0, dataLen)
	active := make(grafana_json.QueryTableResponseNumberColumn, 0, dataLen)
	recovered := make(grafana_json.QueryTableResponseNumberColumn, 0, dataLen)
	deaths := make(grafana_json.QueryTableResponseNumberColumn, 0, dataLen)

	for _, entry := range resp.Series {
		timestamps = append(timestamps, entry.Timestamp)
		active = append(active, float64(entry.Confirmed-entry.Recovered-entry.Deaths))
		recovered = append(recovered, float64(entry.Recovered))
		deaths = append(deaths, float64(entry.Deaths))
	}

	response = new(grafana_json.QueryTableResponse)
	response.Columns = []grafana_json.QueryTableResponseColumn{
		{Text: "timestamp", Data: timestamps},
		{Text: "active", Data: active},
		{Text: "recovered", Data: recovered},
		{Text: "deaths", Data: deaths},
	}
	return
}
