package apiserver

import (
	"log"
	"time"
	"io"
	"fmt"
	"sort"
	"net/http"
	"encoding/json"

	"github.com/gorilla/mux"
	simplejson "github.com/bitly/go-simplejson"

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
	w.WriteHeader(http.StatusOK)
	targetsJson, _ := json.Marshal(apiserver.apihandler.search())
	w.Write(targetsJson)
}

type RequestParameters struct {
	MaxDataPoints int
	From time.Time
	To time.Time
	Targets []string
}

func parseRequest(body io.Reader, validTargets []string) (*RequestParameters, error) {
	parameters := new(RequestParameters)
	js, err := simplejson.NewFromReader(body)

	if err != nil {
		return parameters, err
	}

	parameters.MaxDataPoints = js.Get("maxDataPoints").MustInt()
	parameters.From, err     = time.Parse("2006-01-02", js.Get("range").Get("from").MustString())
	parameters.To, err       = time.Parse("2006-01-02", js.Get("range").Get("to").MustString())
	for i := 0; i < len(js.Get("targets").MustArray()); i++ {
		target := js.Get("targets").GetIndex(i).Get("target").MustString()
		if sort.SearchStrings(validTargets, target) > 0 {
			parameters.Targets = append(parameters.Targets, target)
		} else {
			log.Printf("Unsupported target: '%s'. Dropping", target)
		}
	}

	return parameters, err
}

func (apiserver *APIServer) query(w http.ResponseWriter, req *http.Request) {
	parameters, err := parseRequest(req.Body, apiserver.apihandler.search())

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	output, err := apiserver.apihandler.query(parameters)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	targetsJson, _ := json.Marshal(output)
	w.Write(targetsJson)
}

func (apiserver *APIServer) Run() {
		r := mux.NewRouter()
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
	Target string
	Datapoints [][]int64
}

func (apihandler *APIHandler) query(params *RequestParameters) ([]series, error) {
	entries, err := apihandler.db.List()

	if err != nil {
		return make([]series, 0), err
	}

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


