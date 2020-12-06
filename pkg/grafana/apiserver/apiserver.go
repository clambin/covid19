package apiserver

import (
	"io/ioutil"
	"bytes"

	"time"
	"io"
	"fmt"
	"net/http"
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log        "github.com/sirupsen/logrus"
)

// APIServer implements a generic frameworks for the Grafana simpleJson API datasource 
type APIServer struct {
	apihandler APIHandler
	port       int
}

// Create creates a APIServer object
func Create(apihandler APIHandler, port int) (*APIServer) {
	return &APIServer{apihandler: apihandler, port: port}
}

// APIHandler implements the business logic of the Grafana API datasource so that 
// APIServer can be limited to providing the generic search/query framework
type APIHandler interface{
	Search()                ([]string)
	Query(*APIQueryRequest) (*APIQueryResponse, error)
}

// APIQueryRequest contains the request parameters to the API's 'query' method
type APIQueryRequest struct {
	Range struct {
		From time.Time
		To   time.Time
		// Raw  map[string]string
	}
	Targets []struct{Target string}
}

// APIQueryResponse contains the response of the API's 'query' method
type APIQueryResponse struct {
    Target string           `json:"target"`
    Datapoints [][]int64    `json:"datapoints"`
}
// Run the API Server
func (apiserver *APIServer) Run() {
	r := mux.NewRouter()
	r.Use(prometheusMiddleware)
	r.Path("/metrics").Handler(promhttp.Handler())
	r.HandleFunc("/", apiserver.hello)
	r.HandleFunc("/search", apiserver.search).Methods("POST")
	r.HandleFunc("/query", apiserver.query).Methods("POST")

	http.ListenAndServe(fmt.Sprintf(":%d", apiserver.port), r)
}

// Implement three endpoints. /search and /query are used by Grafana's simple json API datasource
// We also implement / to be used as a K8s liveness probe
//
// Code in these functions is generic. Business logic is provided by the APIHandler object

func (apiserver *APIServer) hello(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello")
}

func (apiserver *APIServer) search(w http.ResponseWriter, req *http.Request) {
	log.Info("/search")
	output := apiserver.apihandler.Search()
	log.Debugf("/search: '%s'", output)
	targetsJSON, _ := json.Marshal(output)
	w.WriteHeader(http.StatusOK)
	w.Write(targetsJSON)
}

func parseRequest(body io.Reader, validTargets []string) (*APIQueryRequest, error) {
	buf, bodyErr := ioutil.ReadAll(body)
	if bodyErr != nil {
		log.Warningf("bodyErr %s", bodyErr.Error())
		return nil, bodyErr
	}
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
	log.Debugf("BODY: %q", rdr1)
	body = rdr2

	var params APIQueryRequest
	decoder := json.NewDecoder(body)
	err := decoder.Decode(&params)
	return &params, err
}

func (apiserver *APIServer) query(w http.ResponseWriter, req *http.Request) {
	log.Info("/query")

	/*
	f, err := os.Create("query.prof")
	if err != nil {
		log.Fatal(err)
	}

	err = pprof.StartCPUProfile(f)
	if err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	*/

	parameters, err := parseRequest(req.Body, apiserver.apihandler.Search())

	if err != nil {
		log.Warningf("error parsing the request (%v). Aborting", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	output, err := apiserver.apihandler.Query(parameters)

	if err != nil {
		log.Warning("Internal Server Error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debugf("/query: %v", output)
	w.WriteHeader(http.StatusOK)
	targetsJSON, _ := json.Marshal(output)
	w.Write(targetsJSON)
	/*
	pprof.StopCPUProfile()
	f.Close()
	*/
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

