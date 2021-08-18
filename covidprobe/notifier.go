package covidprobe

import (
	"fmt"
	"github.com/clambin/covid19/coviddb"
	log "github.com/sirupsen/logrus"
)

// Notifier will send notifications when we receive new updates for selected countries
type Notifier struct {
	Sender        NotificationSender
	lastDBEntries map[string]coviddb.CountryEntry
}

// NewNotifier creates a new Notifier
func NewNotifier(sender NotificationSender, countries []string, db coviddb.DB) *Notifier {
	// TODO: make this a DB function?
	lastDBEntries := make(map[string]coviddb.CountryEntry)

	for _, country := range countries {
		entry, ok, err := db.GetLastForCountry(country)

		if err != nil {
			log.WithError(err).Fatal("unable to access database")
		}

		if ok == false {
			entry = &coviddb.CountryEntry{}
		}

		lastDBEntries[country] = *entry
	}

	return &Notifier{
		Sender:        sender,
		lastDBEntries: lastDBEntries,
	}
}

// Notify sends notification when we receive updates for selected countries
func (notifier *Notifier) Notify(records []coviddb.CountryEntry) (err error) {
	for _, record := range records {
		lastDBEntry, ok := notifier.lastDBEntries[record.Name]

		if ok == false {
			continue
		}

		if record.Timestamp.After(lastDBEntry.Timestamp) {
			title := "New covid data for " + record.Name

			message := fmt.Sprintf("Confirmed: %d, deaths: %d, recovered: %d",
				record.Confirmed-lastDBEntry.Confirmed,
				record.Deaths-lastDBEntry.Deaths,
				record.Recovered-lastDBEntry.Recovered,
			)

			err = notifier.Sender.Send(title, message)

			if err != nil {
				log.WithError(err).Warning("failed to send notification")
				return
			}

			notifier.lastDBEntries[record.Name] = record
		}
	}
	return
}
