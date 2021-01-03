// package reporters handles all reporting needs to other systems
package reporters

import (
	"covid19/internal/coviddb"
)

// Reporter interface for reporting functions
type Reporter interface {
	Report([]coviddb.CountryEntry)
}

// Reporters structure
type Reporters struct {
	reporters []Reporter
}

// Create a new Reporters object
func Create() *Reporters {
	return &Reporters{reporters: make([]Reporter, 0)}
}

func (reporters *Reporters) Add(reporter Reporter) {
	reporters.reporters = append(reporters.reporters, reporter)
}

func (reporters *Reporters) Report(entries []coviddb.CountryEntry) {
	for _, reporter := range reporters.reporters {
		reporter.Report(entries)
	}
}
