package covidhandler

import (
	"errors"
	"fmt"
	"github.com/clambin/covid19/covidcache"
	"github.com/clambin/grafana-json"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

// CovidHandler implements business logic for APIServer
type CovidHandler struct {
	cache *covidcache.Cache
}

// Create a CovidAPIHandler object
func Create(cache *covidcache.Cache) (*CovidHandler, error) {
	if cache == nil {
		return nil, errors.New("no database specified")
	}

	return &CovidHandler{cache: cache}, nil
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

// Endpoints tells the server which endpoints we have implemented
func (handler *CovidHandler) Endpoints() grafana_json.Endpoints {
	return grafana_json.Endpoints{
		Search:     handler.Search,
		Query:      handler.Query,
		TableQuery: handler.TableQuery,
	}
}

// Search returns all supported targets
func (handler *CovidHandler) Search() []string {
	return Targets
}

// Query the DB and return the requested targets
func (handler *CovidHandler) Query(target string, args *grafana_json.TimeSeriesQueryArgs) (response *grafana_json.QueryResponse, err error) {
	start := time.Now()

	deltas := false
	subTarget := target
	if strings.HasSuffix(target, "-delta") {
		deltas = true
		subTarget = strings.TrimSuffix(target, "-delta")
	}

	var resp []covidcache.CacheEntry
	if deltas == false {
		resp = handler.cache.GetTotals(args.Range.To)
	} else {
		resp = handler.cache.GetDeltas(args.Range.To)
	}

	response = new(grafana_json.QueryResponse)
	response.Target = target
	response.DataPoints = make([]grafana_json.QueryResponseDataPoint, len(resp))

loop:
	for index, entry := range resp {
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

	log.WithFields(log.Fields{
		"target": target,
		"time":   time.Now().Sub(start).String(),
		"count":  len(response.DataPoints),
	}).Info("query timeserie")

	return
}

func (handler *CovidHandler) TableQuery(target string, args *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
	start := time.Now()

	switch target {
	case "daily":
		response = handler.buildDaily(args)
	case "cumulative":
		response = handler.buildCumulative(args)
	default:
		err = fmt.Errorf("%s does not implement table query", target)
	}

	var dataLen int
	if response != nil && len(response.Columns) > 0 {
		switch data := response.Columns[0].Data.(type) {
		case grafana_json.TableQueryResponseTimeColumn:
			dataLen = len(data)
		}
	}

	log.WithFields(log.Fields{
		"target": target,
		"time":   time.Now().Sub(start),
		"count":  dataLen,
	}).Info("query table")

	return
}

func (handler *CovidHandler) buildDaily(args *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse) {
	resp := handler.cache.GetDeltas(args.Range.To)

	dataLen := len(resp)
	response = new(grafana_json.TableQueryResponse)
	timestamps := make(grafana_json.TableQueryResponseTimeColumn, 0, dataLen)
	confirmed := make(grafana_json.TableQueryResponseNumberColumn, 0, dataLen)
	recovered := make(grafana_json.TableQueryResponseNumberColumn, 0, dataLen)
	deaths := make(grafana_json.TableQueryResponseNumberColumn, 0, dataLen)

	for _, entry := range resp {
		timestamps = append(timestamps, entry.Timestamp)
		confirmed = append(confirmed, float64(entry.Confirmed))
		recovered = append(recovered, float64(entry.Recovered))
		deaths = append(deaths, float64(entry.Deaths))
	}

	response = new(grafana_json.TableQueryResponse)
	response.Columns = []grafana_json.TableQueryResponseColumn{
		{Text: "timestamp", Data: timestamps},
		{Text: "confirmed", Data: confirmed},
		{Text: "recovered", Data: recovered},
		{Text: "deaths", Data: deaths},
	}
	return
}

func (handler *CovidHandler) buildCumulative(args *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse) {
	resp := handler.cache.GetTotals(args.Range.To)

	dataLen := len(resp)
	response = new(grafana_json.TableQueryResponse)
	timestamps := make(grafana_json.TableQueryResponseTimeColumn, 0, dataLen)
	active := make(grafana_json.TableQueryResponseNumberColumn, 0, dataLen)
	recovered := make(grafana_json.TableQueryResponseNumberColumn, 0, dataLen)
	deaths := make(grafana_json.TableQueryResponseNumberColumn, 0, dataLen)

	for _, entry := range resp {
		timestamps = append(timestamps, entry.Timestamp)
		active = append(active, float64(entry.Confirmed-entry.Recovered-entry.Deaths))
		recovered = append(recovered, float64(entry.Recovered))
		deaths = append(deaths, float64(entry.Deaths))
	}

	response = new(grafana_json.TableQueryResponse)
	response.Columns = []grafana_json.TableQueryResponseColumn{
		{Text: "timestamp", Data: timestamps},
		{Text: "active", Data: active},
		{Text: "deaths", Data: deaths},
		{Text: "recovered", Data: recovered},
	}
	return
}
