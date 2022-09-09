package notifier

import (
	"fmt"
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/models"
	log "github.com/sirupsen/logrus"
)

// Notifier will send notifications when we receive new updates for selected countries
//
//go:generate mockery --name Notifier
type Notifier interface {
	Notify(entries []models.CountryEntry) (err error)
}

// RealNotifier will send notifications when we receive new updates for selected countries
type RealNotifier struct {
	Sender        NotificationSender
	lastDBEntries map[string]models.CountryEntry
}

var _ Notifier = &RealNotifier{}

// NewNotifier creates a new RealNotifier
func NewNotifier(sender NotificationSender, countries []string, db db.CovidStore) *RealNotifier {
	lastDBEntries := make(map[string]models.CountryEntry)

	entries, err := db.GetLatestForCountries(countries)
	if err != nil {
		log.WithError(err).Fatal("unable to access database")
	}

	for name, entry := range entries {
		lastDBEntries[name] = entry
	}

	return &RealNotifier{
		Sender:        sender,
		lastDBEntries: lastDBEntries,
	}
}

// Notify sends notification when we receive updates for selected countries
func (notifier *RealNotifier) Notify(entries []models.CountryEntry) (err error) {
	for _, record := range entries {
		lastDBEntry, ok := notifier.lastDBEntries[record.Name]

		if !ok || !record.Timestamp.After(lastDBEntry.Timestamp) {
			continue
		}

		if record.Confirmed == lastDBEntry.Confirmed &&
			record.Deaths == lastDBEntry.Deaths &&
			record.Recovered == lastDBEntry.Recovered {
			continue
		}

		err = notifier.Sender.Send(
			"New probe data for "+record.Name,
			fmt.Sprintf("Confirmed: %d, deaths: %d, recovered: %d",
				record.Confirmed-lastDBEntry.Confirmed,
				record.Deaths-lastDBEntry.Deaths,
				record.Recovered-lastDBEntry.Recovered,
			),
		)

		if err != nil {
			err = fmt.Errorf("failed to send notifications: " + err.Error())
			break
		}

		notifier.lastDBEntries[record.Name] = record
	}
	return
}
