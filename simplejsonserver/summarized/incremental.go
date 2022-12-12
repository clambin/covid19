package summarized

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v5"
)

// IncrementalHandler returns the incremental number of cases & deaths. If an adhoc filter exists, it returns the
// incremental cases/deaths for that country
type IncrementalHandler struct {
	Retriever
}

var _ simplejson.Handler = &IncrementalHandler{}

func (handler IncrementalHandler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		Query:     handler.tableQuery,
		TagKeys:   handler.tagKeys,
		TagValues: handler.tagValues,
	}
}

func (handler *IncrementalHandler) tableQuery(_ context.Context, req simplejson.QueryRequest) (response simplejson.Response, err error) {
	var entries []models.CountryEntry
	if len(req.Args.AdHocFilters) > 0 {
		entries, err = handler.Retriever.getTotalsForCountry(req.QueryArgs)
	} else {
		entries, err = handler.Retriever.DB.GetTotalsPerDay()
	}

	if err == nil {
		response = createDeltas(dbEntriesToTable(entries)).Filter(req.Args).CreateTableResponse()
	}
	return
}

func (handler *IncrementalHandler) tagKeys(_ context.Context) []string {
	return []string{"Country Name"}
}

func (handler *IncrementalHandler) tagValues(_ context.Context, key string) (values []string, err error) {
	if key != "Country Name" {
		return values, fmt.Errorf("unsupported tag '%s'", key)
	}

	return handler.Retriever.DB.GetAllCountryNames()
}
