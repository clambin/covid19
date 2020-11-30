package grafana

import (
	"log"
	"time"
	"io"
	"fmt"
	"net/http"
	"encoding/json"

	"github.com/gorilla/mux"
	simplejson "github.com/bitly/go-simplejson"
)

var (
	targets = []string{
		"confirmed",
		"confirmed-delta",
		"death",
		"death-delta",
		"recovered",
		"recovered-delta",
		"active", "active-delta",
	}
)

func hello(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello")
}

func search(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	targetsJson, _ := json.Marshal(targets)
	w.Write(targetsJson)
}

// {
//	"maxDataPoints": 10,
//	"interval": "1y",
//	"range": {
//		"from": "2020-11-01",
//		"to": "2020-11-02"
//	},
//	"targets": [
//		{ "target": "death", "type": "foo" }
//	]
// }

type GrafanaRequestParameters struct {
	MaxDataPoints int
	From time.Time
	To time.Time
	Targets []string
}

func parseRequest(body io.Reader) (GrafanaRequestParameters, error) {
	parameters := GrafanaRequestParameters{}
	js, err := simplejson.NewFromReader(body)

	if err != nil {
		return parameters, err
	}

	parameters.MaxDataPoints = js.Get("maxDataPoints").MustInt()
	parameters.From, err     = time.Parse("2006-01-02", js.Get("range").Get("from").MustString())
	parameters.To, err       = time.Parse("2006-01-02", js.Get("range").Get("to").MustString())
	for i := 0; i < len(js.Get("targets").MustArray()); i++ {
		parameters.Targets = append(
				parameters.Targets, js.Get("targets").GetIndex(i).Get("target").MustString())
	}

	return parameters, err
}

func query(w http.ResponseWriter, req *http.Request) {
	parameters, err := parseRequest(req.Body)

	if err == nil {
		log.Print(parameters)
	}
}

func Run() {
		r := mux.NewRouter()
		r.HandleFunc("/", hello)
		r.HandleFunc("/search", search).Methods("POST")
		r.HandleFunc("/query", query).Methods("POST")

		http.ListenAndServe(":5000", r)
}
