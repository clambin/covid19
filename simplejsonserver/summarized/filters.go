package summarized

import (
	"fmt"
	"github.com/clambin/simplejson/v5"
)

func evaluateAdHocFilter(adHocFilters []simplejson.AdHocFilter) (name string, err error) {
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
