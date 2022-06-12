package summarized

import (
	"context"
	"fmt"
	"github.com/clambin/covid19/models"
	"github.com/clambin/simplejson/v3"
	"github.com/clambin/simplejson/v3/query"
)

// CumulativeHandler returns the incremental number of cases & deaths. If an adhoc filter exists, it returns the
// cumulative cases/deaths for that country
type CumulativeHandler struct {
	Retriever
}

var _ simplejson.Handler = &CumulativeHandler{}

func (handler CumulativeHandler) Endpoints() (endpoints simplejson.Endpoints) {
	return simplejson.Endpoints{
		Query:     handler.tableQuery,
		TagKeys:   handler.tagKeys,
		TagValues: handler.tagValues,
	}
}

func (handler *CumulativeHandler) tableQuery(_ context.Context, req query.Request) (response query.Response, err error) {
	var entries []models.CountryEntry
	if len(req.Args.AdHocFilters) > 0 {
		entries, err = handler.Retriever.getTotalsForCountry(req.Args)
	} else {
		entries, err = handler.Retriever.DB.GetTotalsPerDay()
	}

	if err == nil {
		response = dbEntriesToTable(entries).Filter(req.Args).CreateTableResponse()
	}
	return
}

func (handler *CumulativeHandler) tagKeys(_ context.Context) []string {
	return []string{"Country Name"}
}

func (handler *CumulativeHandler) tagValues(_ context.Context, key string) (values []string, err error) {
	if key != "Country Name" {
		return values, fmt.Errorf("unsupported tag '%s'", key)
	}

	return handler.Retriever.DB.GetAllCountryNames()
}
