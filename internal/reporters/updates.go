package reporters

import (
	"covid19/internal/coviddb"

	"github.com/arcanericky/pushover"
	slack "github.com/ashwanthkumar/slack-go-webhook"
	log "github.com/sirupsen/logrus"

	"fmt"
)

type ReportsConfiguration struct {
	Countries []string
	Updates   struct {
		Pushover struct {
			Token string
			User  string
		}
		Slack struct {
			URL     string
			Channel string
		}
	}
}

// UpdatesReporter reports new data for a list of countries via pushover
type UpdatesReporter struct {
	config   *ReportsConfiguration
	db       coviddb.DB
	SentReqs []string
}

// NewNewUpdatesReporter creates a new NewUpdatesReporter object for the specified list of countries.
// Selected countries getting new data will be reported via Pushover with the specified
// Pushover API & User tokens.
func NewUpdatesReporter(config *ReportsConfiguration, db coviddb.DB) *UpdatesReporter {
	return &UpdatesReporter{
		config:   config,
		db:       db,
		SentReqs: make([]string, 0),
	}
}

// Report acts on new covidprobe entries for countries we want to report on
func (reporter *UpdatesReporter) Report(entries []coviddb.CountryEntry) {
	toReport := make([]coviddb.CountryEntry, 0)
	for _, entry := range entries {
		if isSelected(entry.Name, reporter.config.Countries) {
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
				title := fmt.Sprintf("New covid data for %s", entry.Name)
				message := fmt.Sprintf("New confirmed: %d\nNew deaths: %d\nNew recovered: %d",
					entry.Confirmed-dbEntry.Confirmed,
					entry.Deaths-dbEntry.Deaths,
					entry.Recovered-dbEntry.Recovered,
				)
				reporter.SentReqs = append(reporter.SentReqs, message)

				if reporter.config.Updates.Pushover.Token != "" && reporter.config.Updates.Pushover.User != "" {
					_, err = pushover.Message(pushover.MessageRequest{
						Token:   reporter.config.Updates.Pushover.Token,
						User:    reporter.config.Updates.Pushover.User,
						Title:   title,
						Message: message,
					},
					)
				}
				if reporter.config.Updates.Slack.URL != "" {
					payload := slack.Payload{
						Channel: reporter.config.Updates.Slack.Channel,
						Text:    title + "\n" + message,
					}
					_ = slack.Send(reporter.config.Updates.Slack.URL, "", payload)
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
