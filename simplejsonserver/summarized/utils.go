package summarized

import (
	"fmt"
	"github.com/clambin/covid19/cache"
	"github.com/clambin/simplejson/v2/common"
	"github.com/clambin/simplejson/v2/query"
	"time"
)

func buildResponse(entries []cache.Entry, window common.Range) *query.TableResponse {
	timestamps := make([]time.Time, 0, len(entries))
	confirmed := make([]float64, 0, len(entries))
	deaths := make([]float64, 0, len(entries))

	for _, entry := range entries {
		if entry.Timestamp.Before(window.From) {
			continue
		}
		if entry.Timestamp.After(window.To) {
			break
		}
		timestamps = append(timestamps, entry.Timestamp)
		confirmed = append(confirmed, float64(entry.Confirmed))
		deaths = append(deaths, float64(entry.Deaths))
	}

	return &query.TableResponse{
		Columns: []query.Column{
			{Text: "timestamp", Data: query.TimeColumn(timestamps)},
			{Text: "deaths", Data: query.NumberColumn(deaths)},
			{Text: "confirmed", Data: query.NumberColumn(confirmed)},
		},
	}
}

func evaluateAdHocFilter(adHocFilters []common.AdHocFilter) (name string, err error) {
	if len(adHocFilters) != 1 {
		err = fmt.Errorf("only one ad hoc filter supported. got %d", len(adHocFilters))
	} else if adHocFilters[0].Key != "Country Name" {
		err = fmt.Errorf("only \"Country Name\" is supported in ad hoc filter. got %s", adHocFilters[0].Key)
	} else if adHocFilters[0].Operator != "=" {
		err = fmt.Errorf("only \"=\" operator supported in ad hoc filter. got %s", adHocFilters[0].Operator)
	} else {
		name = adHocFilters[0].Value
	}
	return
}
