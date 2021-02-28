package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"time"
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
	Range         APIQueryRequestRange    `json:"range"`
	Interval      string                  `json:"interval"`
	MaxDataPoints uint64                  `json:"maxDataPoints"`
	Targets       []APIQueryRequestTarget `json:"targets"`
}

type APIQueryRequestRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type APIQueryRequestTarget struct {
	Target string `json:"target"`
	Type   string `json:"type"`
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
	output := apiServer.apiHandler.Search()
	log.WithField("output", output).Debug("/search")
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
	var (
		err        error
		parameters *APIQueryRequest
		response   []APIQueryResponse
	)
	defer req.Body.Close()

	if parameters, err = parseRequest(req.Body); err == nil {
		if response, err = apiServer.apiHandler.Query(parameters); err == nil {
			maxPoints := 0
			for _, item := range response {
				if len(item.DataPoints) > maxPoints {
					maxPoints = len(item.DataPoints)
				}
			}
			log.WithFields(log.Fields{
				"maxDataPoints":    parameters.MaxDataPoints,
				"interval":         parameters.Interval,
				"actualDataPoints": maxPoints,
			}).Debug("/query")
			w.WriteHeader(http.StatusOK)
			targetsJSON, _ := json.Marshal(response)
			w.Write(targetsJSON)
		} else {
			log.WithField("err", err).Warning("Internal Server Error")
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		log.WithField("err", err).Warning("error parsing request")
		w.WriteHeader(http.StatusBadRequest)
	}

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
