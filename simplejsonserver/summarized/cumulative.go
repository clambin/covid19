package summarized

import (
	"context"
	"fmt"
	"github.com/clambin/simplejson/v6"
)

// CumulativeHandler returns the incremental number of cases & deaths. If an adhoc filter exists, it returns the
// cumulative cases/deaths for that country
type CumulativeHandler struct {
	Fetcher
}

var _ simplejson.Handler = &CumulativeHandler{}

func (handler *CumulativeHandler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		Query:     handler.tableQuery,
		TagKeys:   handler.tagKeys,
		TagValues: handler.tagValues,
	}
}

func (handler *CumulativeHandler) tableQuery(_ context.Context, req simplejson.QueryRequest) (simplejson.Response, error) {
	entries, err := handler.Fetcher.getTotals(req.QueryArgs)
	if err != nil {
		return nil, err
	}
	return dbEntriesToTable(entries).Filter(req.QueryArgs.Args).CreateTableResponse(), nil
}

func (handler *CumulativeHandler) tagKeys(_ context.Context) []string {
	return []string{"Country Name"}
}

func (handler *CumulativeHandler) tagValues(_ context.Context, key string) (values []string, err error) {
	if key != "Country Name" {
		return values, fmt.Errorf("unsupported tag '%s'", key)
	}

	return handler.Fetcher.DB.GetAllCountryNames()
}
