package summarized

import (
	"context"
	"fmt"
	"github.com/clambin/simplejson/v6"
)

// IncrementalHandler returns the incremental number of cases & deaths. If an adhoc filter exists, it returns the
// incremental cases/deaths for that country
type IncrementalHandler struct {
	Fetcher
}

var _ simplejson.Handler = &IncrementalHandler{}

func (handler IncrementalHandler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		Query:     handler.tableQuery,
		TagKeys:   handler.tagKeys,
		TagValues: handler.tagValues,
	}
}

func (handler *IncrementalHandler) tableQuery(_ context.Context, req simplejson.QueryRequest) (simplejson.Response, error) {
	entries, err := handler.Fetcher.getTotals(req.QueryArgs)
	if err != nil {
		return nil, err
	}
	return createDeltas(dbEntriesToTable(entries)).Filter(req.Args).CreateTableResponse(), nil
}

func (handler *IncrementalHandler) tagKeys(_ context.Context) []string {
	return []string{"Country Name"}
}

func (handler *IncrementalHandler) tagValues(_ context.Context, key string) (values []string, err error) {
	if key != "Country Name" {
		return values, fmt.Errorf("unsupported tag '%s'", key)
	}

	return handler.Fetcher.DB.GetAllCountryNames()
}
