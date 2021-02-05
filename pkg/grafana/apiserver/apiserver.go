package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// APIServer implements a generic frameworks for the Grafana simpleJson API datasource
type APIServer struct {
	apiHandler APIHandler
	port       int
}

// Create creates a APIServer object
func Create(apiHandler APIHandler, port int) *APIServer {
	return &APIServer{apiHandler: apiHandler, port: port}
}

// APIHandler implements the business logic of the Grafana API datasource so that
// APIServer can be limited to providing the generic search/query framework
type APIHandler interface {
	Search() []string
	Query(*APIQueryRequest) ([]APIQueryResponse, error)
}

// APIQueryRequest contains the request parameters to the API's 'query' method
type APIQueryRequest struct {
	Range struct {
		From time.Time
		To   time.Time
		// Raw  map[string]string
	}
	Targets []struct{ Target string }
}

// APIQueryResponse contains the response of the API's 'query' method
type APIQueryResponse struct {
	Target     string     `json:"target"`
	DataPoints [][2]int64 `json:"datapoints"`
}

// Run the API Server
func (apiServer *APIServer) Run() error {
	r := mux.NewRouter()
	r.Use(prometheusMiddleware)
	r.Path("/metrics").Handler(promhttp.Handler())
	r.HandleFunc("/", apiServer.hello)
	r.HandleFunc("/search", apiServer.search).Methods("POST")
	r.HandleFunc("/query", apiServer.query).Methods("POST")

	return http.ListenAndServe(fmt.Sprintf(":%d", apiServer.port), r)
}

// Implement three endpoints. /search and /query are used by Grafana's simple json API datasource
// We also implement / to be used as a K8s liveness probe
//
// Code in these functions is generic. Business logic is provided by the APIHandler object

func (apiServer *APIServer) hello(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Hello")
}

func (apiServer *APIServer) search(w http.ResponseWriter, _ *http.Request) {
	log.Info("/search")
	output := apiServer.apiHandler.Search()
	log.Debugf("/search: '%s'", output)
	targetsJSON, _ := json.Marshal(output)
	w.WriteHeader(http.StatusOK)
	w.Write(targetsJSON)
}

func parseRequest(body io.Reader) (*APIQueryRequest, error) {
	var params APIQueryRequest
	buf, err := ioutil.ReadAll(body)
	if err == nil {
		rdr := ioutil.NopCloser(bytes.NewBuffer(buf))
		decoder := json.NewDecoder(rdr)
		err = decoder.Decode(&params)
	}
	return &params, err
}

func (apiServer *APIServer) query(w http.ResponseWriter, req *http.Request) {
	log.Info("/query")

	defer req.Body.Close()
	parameters, err := parseRequest(req.Body)

	if err != nil {
		log.Warningf("error parsing the request (%v). Aborting", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	output, err := apiServer.apiHandler.Query(parameters)

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
