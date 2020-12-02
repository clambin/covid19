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
)

// GrafanaAPIHandler implements the business logic of the Grafana API datasource so that 
// GrafanaAPIServer can be limited to providing the generic search/queury framework
type GrafanaAPIHandler interface{
	search()                  ([]string)
	// FIXME: best way to make query signature independent from expected output
	query(RequestParameters)  ([]series, error)
}

// GrafanaAPIServer implements a generic frameworks for the Grafana simpleJson API datasource 
type GrafanaAPIServer struct {
	apihandler GrafanaAPIHandler
}

// CreateGrafanaAPIServer creates a GrafanaAPIServer object
func CreateGrafanaAPIServer(apihandler GrafanaAPIHandler) (GrafanaAPIServer) {
	return GrafanaAPIServer{apihandler: apihandler}
}

// Prometheus metrics
var (
  httpDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
    Name: "grafana_api_duration_seconds",
    Help: "Grafana API duration of HTTP requests.",
  }, []string{"path"})
)

func prometheusMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    route := mux.CurrentRoute(r)
    path, _ := route.GetPathTemplate()
    timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
    next.ServeHTTP(w, r)
    timer.ObserveDuration()
  })
}

// Run the API Server
func (apiserver *GrafanaAPIServer) Run() {
		r := mux.NewRouter()
		r.Use(prometheusMiddleware)
		r.Path("/metrics").Handler(promhttp.Handler())
		r.HandleFunc("/", apiserver.hello)
		r.HandleFunc("/search", apiserver.search).Methods("POST")
		r.HandleFunc("/query", apiserver.query).Methods("POST")

		// TODO: use cfg.port
		http.ListenAndServe(":5000", r)
}

// Implement three endpoints. /search and /query are used by Grafana's simple json API datasource
// We also implement / to be used as a K8s liveness probe
//
// Code in these functions is generic. Business logic is provided by GrafanaAPIHandler object

func (apiserver GrafanaAPIServer) hello(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello")
}

func (apiserver GrafanaAPIServer) search(w http.ResponseWriter, req *http.Request) {
	log.Info("/search")
	output := apiserver.apihandler.search()
	log.Debugf("/search: '%s'", output)
	targetsJSON, _ := json.Marshal(output)
	w.WriteHeader(http.StatusOK)
	w.Write(targetsJSON)
}

// RequestParameters contains the (needed) parameters supplied to /query
type RequestParameters struct {
	// MaxDataPoints int
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
	var timestamp string
	parameters := new(RequestParameters)
	js, err := simplejson.NewFromReader(body)

	// if err == nil {
	//	parameters.MaxDataPoints, err  = js.Get("maxDataPoints").Int()
	// }
	if err == nil {
		timestamp, err = js.Get("range").Get("from").String()
		if err == nil {
			parameters.From, err = time.Parse("2006-01-02T15:04:05.000Z", timestamp)
		}
	}
	if err == nil {
		timestamp, err = js.Get("range").Get("to").String()
		if err == nil {
			parameters.To, err = time.Parse("2006-01-02T15:04:05.000Z", timestamp)
		}
	}
	if err == nil {
		for i := 0; i < len(js.Get("targets").MustArray()); i++ {
			target := js.Get("targets").GetIndex(i).Get("target").MustString()
			if isValidTarget(target, validTargets) {
				parameters.Targets = append(parameters.Targets, target)
			} else {
				log.Warningf("Unsupported target: '%s'. Dropping", target)
			}
		}
	}

	return parameters, err
}

func (apiserver *GrafanaAPIServer) query(w http.ResponseWriter, req *http.Request) {
	log.Info("/query")
	parameters, err := parseRequest(req.Body, apiserver.apihandler.search())

	if err != nil {
		log.Warningf("error parsing the request (%v). Aborting", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	output, err := apiserver.apihandler.query(*parameters)

	if err != nil {
		log.Warning("Internal Server Error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debugf("/query: %v", output)
	w.WriteHeader(http.StatusOK)
	targetsJSON, _ := json.Marshal(output)
	w.Write(targetsJSON)
}
