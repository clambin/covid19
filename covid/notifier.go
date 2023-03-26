package covid

import (
	"fmt"
	"github.com/clambin/covid19/covid/shoutrrr"
	"github.com/clambin/covid19/models"
	"github.com/clambin/go-common/set"
	"golang.org/x/exp/slog"
)

type Notifier struct {
	shoutrrr.Sender
	Countries set.Set[string]
}

func (n Notifier) Notify(current map[string]models.CountryEntry, updates []models.CountryEntry) error {
	for _, update := range updates {
		if !n.Countries.Contains(update.Name) {
			continue
		}

		entry, ok := current[update.Name]
		if !ok {
			continue
		}

		if !update.Timestamp.After(entry.Timestamp) {
			continue
		}

		slog.Info("update", "confirmed", update.Confirmed-entry.Confirmed, "deaths", update.Deaths-entry.Deaths)

		err := n.Sender.Send(
			"New data for "+update.Name,
			fmt.Sprintf("Confirmed: %d, deaths: %d", update.Confirmed-entry.Confirmed, update.Deaths-entry.Deaths),
		)
		if err != nil {
			slog.Error("failed to send notification", "err", err)
		}
	}
	return nil
}
