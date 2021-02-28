package covidhandler

import (
	"covid19/internal/covidcache"
	"errors"
	log "github.com/sirupsen/logrus"

	"covid19/pkg/grafana/apiserver"
)

// APIHandler implements business logic for APIServer
type APIHandler struct {
	cache *covidcache.Cache
}

// Create a CovidAPIHandler object
func Create(cache *covidcache.Cache) (*APIHandler, error) {
	if cache == nil {
		return nil, errors.New("no database specified")
	}
	return &APIHandler{cache: cache}, nil
}

var (
	targets = []string{
		"active",
		"active-delta",
		"confirmed",
		"confirmed-delta",
		"death",
		"death-delta",
		"recovered",
		"recovered-delta",
	}
)

// Search returns all supported targets
func (apiHandler *APIHandler) Search() []string {
	return targets
}

// Query the DB and return the requested targets
func (apiHandler *APIHandler) Query(request *apiserver.APIQueryRequest) (response []apiserver.APIQueryResponse, err error) {
	totals := apiHandler.cache.GetTotals(request.Range.To)
	deltas := apiHandler.cache.GetDeltas(request.Range.To)

	for _, target := range request.Targets {
		switch target.Target {
		case "confirmed":
			response = append(response, buildResponsePart(totals, target.Target, "confirmed"))
		case "confirmed-delta":
			response = append(response, buildResponsePart(deltas, target.Target, "confirmed"))
		case "recovered":
			response = append(response, buildResponsePart(totals, target.Target, "recovered"))
		case "recovered-delta":
			response = append(response, buildResponsePart(deltas, target.Target, "recovered"))
		case "death":
			response = append(response, buildResponsePart(totals, target.Target, "death"))
		case "death-delta":
			response = append(response, buildResponsePart(deltas, target.Target, "death"))
		case "active":
			response = append(response, buildResponsePart(totals, target.Target, "active"))
		case "active-delta":
			response = append(response, buildResponsePart(deltas, target.Target, "active"))
		default:
			log.Warningf("dropping unsupported target: %s", target.Target)
		}
	}
	return
}

func buildResponsePart(entries []covidcache.CacheEntry, target string, attribute string) (response apiserver.APIQueryResponse) {
	var timestamp, value int64

	response.Target = target
	response.DataPoints = make([][2]int64, 0)
	for _, entry := range entries {
		timestamp = entry.Timestamp.UnixNano() / 1000000
		value = 0
		switch attribute {
		case "confirmed":
			value = entry.Confirmed
		case "recovered":
			value = entry.Recovered
		case "death":
			value = entry.Deaths
		case "active":
			value = entry.Active
		}
		response.DataPoints = append(response.DataPoints, [2]int64{value, timestamp})
	}
	return
}
