package notifier

import (
	"fmt"
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/models"
	"github.com/clambin/go-common/set"
)

// Notifier sends notifications when we receive new updates for selected countries
type Notifier struct {
	Router        Router
	lastDBEntries map[string]models.CountryEntry
}

// NewNotifier creates a new RealNotifier
func NewNotifier(r Router, countries []string, db db.CovidStore) (*Notifier, error) {
	entries, err := db.GetLatestForCountries()
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	uniqueCountries := set.Create(countries...)
	lastDBEntries := make(map[string]models.CountryEntry)
	for name, entry := range entries {
		if uniqueCountries.Contains(name) {
			lastDBEntries[name] = entry
		}
	}

	return &Notifier{Router: r, lastDBEntries: lastDBEntries}, nil
}

func (n *Notifier) Notify(entries []models.CountryEntry) (err error) {
	for _, record := range entries {
		lastDBEntry, ok := n.lastDBEntries[record.Name]

		if !ok || !record.Timestamp.After(lastDBEntry.Timestamp) {
			continue
		}

		if record.Confirmed == lastDBEntry.Confirmed &&
			record.Deaths == lastDBEntry.Deaths &&
			record.Recovered == lastDBEntry.Recovered {
			continue
		}

		err = n.Router.Send(
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

		n.lastDBEntries[record.Name] = record
	}
	return
}
