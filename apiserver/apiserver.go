package apiserver

import (
	"time"
	"io"
	"fmt"
	"net/http"
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	simplejson "github.com/bitly/go-simplejson"
	log        "github.com/sirupsen/logrus"

	"covid19api/coviddb"
)

// Server

type APIServer struct {
	apihandler *APIHandler
}

func Server(apihandler *APIHandler) (apiserver *APIServer) {
	return &APIServer{apihandler: apihandler}
}

func (apiserver *APIServer) hello(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello")
}

func (apiserver *APIServer) search(w http.ResponseWriter, req *http.Request) {
	output := apiserver.apihandler.search()
	log.Debugf("/search: '%s'", output)
	targetsJson, _ := json.Marshal(output)
	w.WriteHeader(http.StatusOK)
	w.Write(targetsJson)
}

type RequestParameters struct {
	MaxDataPoints int
	From time.Time
	To time.Time
	Targets []string
}

func isValidTarget(target string, validTargets []string) (bool) {
	for _, t := range validTargets {
		if t == target {
			return true
		}
	}
	return false
}

func parseRequest(body io.Reader, validTargets []string) (*RequestParameters, error) {
	parameters := new(RequestParameters)
	js, err := simplejson.NewFromReader(body)

	if err != nil {
		return parameters, err
	}

	parameters.MaxDataPoints = js.Get("maxDataPoints").MustInt()
	parameters.From, _       = time.Parse("2006-01-02T15:04:05.000Z", js.Get("range").Get("from").MustString())
	parameters.To, _         = time.Parse("2006-01-02T15:04:05.000Z", js.Get("range").Get("to").MustString())

	for i := 0; i < len(js.Get("targets").MustArray()); i++ {
		target := js.Get("targets").GetIndex(i).Get("target").MustString()
		if isValidTarget(target, validTargets) {
			parameters.Targets = append(parameters.Targets, target)
		} else {
			log.Warningf("Unsupported target: '%s'. Dropping", target)
		}
	}

	return parameters, nil
}

func (apiserver *APIServer) query(w http.ResponseWriter, req *http.Request) {
	log.Info("/query")
	parameters, err := parseRequest(req.Body, apiserver.apihandler.search())

	if err != nil {
		log.Debugf("error parsing the request (%v). Aborting", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Debugf("parameters: %v", parameters)

	output, err := apiserver.apihandler.query(parameters)

	if err != nil {
		log.Debug("Internal Server Error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debugf("/query: %v", output)
	w.WriteHeader(http.StatusOK)
	targetsJson, _ := json.Marshal(output)
	w.Write(targetsJson)
}

var (
  httpDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
    Name: "covid19_http_duration_seconds",
    Help: "Duration of HTTP requests.",
  }, []string{"path"})
)

// prometheusMiddleware implements mux.MiddlewareFunc.
func prometheusMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    route := mux.CurrentRoute(r)
    path, _ := route.GetPathTemplate()
    timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
    next.ServeHTTP(w, r)
    timer.ObserveDuration()
  })
}


func (apiserver *APIServer) Run() {
		r := mux.NewRouter()
		r.Use(prometheusMiddleware)
		r.Path("/metrics").Handler(promhttp.Handler())
		r.HandleFunc("/", apiserver.hello)
		r.HandleFunc("/search", apiserver.search).Methods("POST")
		r.HandleFunc("/query", apiserver.query).Methods("POST")

		http.ListenAndServe(":5000", r)
}

// Handler

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

type APIHandler struct {
	db *coviddb.CovidDB
}

func Handler(db *coviddb.CovidDB) (*APIHandler) {
	return &APIHandler{db: db}
}

func (apihandler *APIHandler) search() ([]string) {
	return targets
}

type series struct {
	Target string           `json:"target"`
	Datapoints [][]int64    `json:"datapoints"`
}

func (apihandler *APIHandler) query(params *RequestParameters) ([]series, error) {
	entries, err := apihandler.db.List(params.To)

	if err != nil {
		return make([]series, 0), err
	}

	log.Debugf("Found %d entries in DB", len(entries))

	return buildSeries(entries, params.Targets), nil
}

func buildSeries(entries []coviddb.CountryEntry, targets []string) ([]series) {
	series_list := make([]series, 0)
	totalcases  := coviddb.GetTotalCases(entries)

	for _, target := range targets {
		switch target {
		case "confirmed":
			series_list = append(series_list, series{Target: target, Datapoints: totalcases[coviddb.CONFIRMED]})
		case "confirmed-delta":
			series_list = append(series_list, series{Target: target, Datapoints: coviddb.GetTotalDeltas(totalcases[coviddb.CONFIRMED])})
		case "recovered":
			series_list = append(series_list, series{Target: target, Datapoints: totalcases[coviddb.RECOVERED]})
		case "recovered-delta":
			series_list = append(series_list, series{Target: target, Datapoints: coviddb.GetTotalDeltas(totalcases[coviddb.RECOVERED])})
		case "deaths":
			series_list = append(series_list, series{Target: target, Datapoints: totalcases[coviddb.DEATHS]})
		case "deaths-delta":
			series_list = append(series_list, series{Target: target, Datapoints: coviddb.GetTotalDeltas(totalcases[coviddb.DEATHS])})
		case "active":
			series_list = append(series_list, series{Target: target, Datapoints: totalcases[coviddb.ACTIVE]})
		case "active-delta":
			series_list = append(series_list, series{Target: target, Datapoints: coviddb.GetTotalDeltas(totalcases[coviddb.ACTIVE])})
		}
	}

	return series_list
}


