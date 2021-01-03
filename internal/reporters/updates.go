package reporters

import (
	"covid19/internal/coviddb"
	"fmt"
	"github.com/arcanericky/pushover"
	log "github.com/sirupsen/logrus"
)

// UpdatesReporter reports new data for a list of countries via pushover
type UpdatesReporter struct {
	token     string
	user      string
	countries []string
	db        coviddb.DB
	SentReqs  []pushover.MessageRequest
}

// NewNewUpdatesReporter creates a new NewUpdatesReporter object for the specified list of countries.
// Selected countries getting new data will be reported via Pushover with the specified
// Pushover API & User tokens.
func NewUpdatesReporter(apiToken, userToken string, countries []string, coviddb coviddb.DB) *UpdatesReporter {
	return &UpdatesReporter{
		token:     apiToken,
		user:      userToken,
		countries: countries,
		db:        coviddb,
		SentReqs:  make([]pushover.MessageRequest, 0),
	}
}

// Report acts on new covid entries for countries we want to report on
func (reporter *UpdatesReporter) Report(entries []coviddb.CountryEntry) {
	toReport := make([]coviddb.CountryEntry, 0)
	for _, entry := range entries {
		if isSelected(entry.Name, reporter.countries) {
			toReport = append(toReport, entry)
		}
	}
	if len(toReport) > 0 {
		reporter.process(toReport)
	}
}

// process examines each new entry. If it's more recent, report the differences to pushover
func (reporter *UpdatesReporter) process(entries []coviddb.CountryEntry) {
	var (
		err     error
		entry   coviddb.CountryEntry
		dbEntry *coviddb.CountryEntry
	)

	for _, entry = range entries {
		if dbEntry, err = reporter.db.GetLastBeforeDate(entry.Name, entry.Timestamp); err == nil {
			if dbEntry != nil {
				req := pushover.MessageRequest{
					Token: reporter.token,
					User:  reporter.user,
					Title: "New covid19 data for " + entry.Name,
					Message: fmt.Sprintf("New confirmed: %d\nNew deaths: %d\nNew recovered: %d",
						entry.Confirmed-dbEntry.Confirmed,
						entry.Deaths-dbEntry.Deaths,
						entry.Recovered-dbEntry.Recovered,
					),
				}
				if req.Token != "" && req.User != "" {
					_, err = pushover.Message(req)
				} else {
					reporter.SentReqs = append(reporter.SentReqs, req)
					// unit test mode: record the entry so we can examine the output
				}
			}
		}
		if err != nil {
			log.Warnf("could not report on new entries: " + err.Error())
			break
		}
	}
}

func isSelected(country string, countries []string) bool {
	for _, c := range countries {
		if country == c {
			return true
		}
	}
	return false
}
