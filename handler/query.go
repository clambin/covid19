package handler

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/cache"
	grafana_json "github.com/clambin/grafana-json"
	log "github.com/sirupsen/logrus"
	"time"
)

// TableQuery returns the table response for the provided target`
func (handler *Handler) TableQuery(_ context.Context, target string, args *grafana_json.TableQueryArgs) (response *grafana_json.TableQueryResponse, err error) {
	start := time.Now()

	switch target {
	case "incremental":
		response, err = handler.handleIncremental(args)
	case "cumulative":
		response, err = handler.handleCumulative(args)
	case "evolution":
		response, err = handler.handleEvolution(args)
	case "country-confirmed":
		response, err = handler.handleLatestConfirmedByCountry(args)
	case "country-deaths":
		response, err = handler.handleLatestDeathsByCountry(args)
	case "country-confirmed-population":
		response, err = handler.handleConfirmedByCountryByPopulation(args)
	case "country-deaths-population":
		response, err = handler.handleDeathsByCountryByPopulation(args)
	case "country-deaths-vs-confirmed":
		response, err = handler.handleMortalityVsConfirmed(args)
	case "updates":
		response, err = handler.handleUpdates(args)
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
		"err":    err,
		"target": target,
		"time":   time.Now().Sub(start),
		"count":  dataLen,
	}).Info("table query")

	return
}

func buildResponse(entries []cache.Entry, window grafana_json.QueryRequestRange) *grafana_json.TableQueryResponse {
	var (
		timestamps []time.Time
		confirmed  []float64
		deaths     []float64
	)
	for _, entry := range entries {
		if entry.Timestamp.Before(window.From) {
			continue
		}
		if entry.Timestamp.After(window.To) {
			break
		}
		timestamps = append(timestamps, entry.Timestamp)
		confirmed = append(confirmed, float64(entry.Confirmed))
		deaths = append(deaths, float64(entry.Deaths))
	}

	return &grafana_json.TableQueryResponse{
		Columns: []grafana_json.TableQueryResponseColumn{
			{Text: "timestamp", Data: grafana_json.TableQueryResponseTimeColumn(timestamps)},
			{Text: "deaths", Data: grafana_json.TableQueryResponseNumberColumn(deaths)},
			{Text: "confirmed", Data: grafana_json.TableQueryResponseNumberColumn(confirmed)},
		},
	}
}

func evaluateAhHocFilter(adHocFilters []grafana_json.AdHocFilter) (name string, err error) {
	if len(adHocFilters) != 1 {
		err = fmt.Errorf("only one ad hoc filter supported. got %d", len(adHocFilters))
	} else if adHocFilters[0].Key != "Country Name" {
		err = fmt.Errorf("only \"Country Name\" is supported in ad hoc filter. got %s", adHocFilters[0].Key)
	} else if adHocFilters[0].Operator != "=" {
		err = fmt.Errorf("only \"=\" operator supported in ad hoc filter. got %s", adHocFilters[0].Operator)
	} else {
		name = adHocFilters[0].Value
	}
	return
}
