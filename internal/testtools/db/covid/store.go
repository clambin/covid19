package covid

import (
	"errors"
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/models"
	"github.com/clambin/go-common/set"
	"sort"
	"time"
)

type FakeStore struct {
	Records []models.CountryEntry
	Fail    bool
	NoSort  bool
}

func (f *FakeStore) Add(entries []models.CountryEntry) error {
	if f.Fail {
		return errors.New("fail")
	}
	f.Records = append(f.Records, entries...)
	return nil
}

func (f *FakeStore) GetLatestForCountries(t time.Time) (map[string]models.CountryEntry, error) {
	if f.Fail {
		return nil, errors.New("fail")
	}

	records := make(map[string]models.CountryEntry)
	for _, record := range f.Records {
		if !t.IsZero() && record.Timestamp.After(t) {
			continue
		}
		current := records[record.Name]
		if current.Timestamp.After(record.Timestamp) {
			continue
		}
		records[record.Name] = record
	}
	return records, nil
}

func (f *FakeStore) GetAllForCountryName(s string) ([]models.CountryEntry, error) {
	records := make([]models.CountryEntry, 0, len(f.Records))
	for _, record := range f.Records {
		if record.Name == s {
			records = append(records, record)
		}
	}
	return records, nil
}

func (f *FakeStore) GetAllCountryNames() ([]string, error) {
	countryNames := set.Create[string]()
	for _, record := range f.Records {
		countryNames.Add(record.Name)
	}
	names := countryNames.List()
	sort.Strings(names)
	return names, nil
}

func (f *FakeStore) GetTotalsPerDay() ([]models.CountryEntry, error) {
	records := make([]models.CountryEntry, len(f.Records))
	copy(records, f.Records)
	if !f.NoSort {
		sort.Slice(records, func(i, j int) bool {
			return records[i].Timestamp.Before(records[j].Timestamp)
		})
	}

	result := make([]models.CountryEntry, 0, len(records))
	row := -1

	for _, record := range records {
		if len(result) == 0 || !result[row].Timestamp.Equal(record.Timestamp) {
			result = append(result, models.CountryEntry{Timestamp: record.Timestamp})
			row++
		}
		result[row].Confirmed += record.Confirmed
		result[row].Recovered += record.Recovered
		result[row].Deaths += record.Deaths
	}

	return result, nil
}

func (f *FakeStore) GetAllForRange(from time.Time, to time.Time) ([]models.CountryEntry, error) {
	var filteredRecords []models.CountryEntry
	for _, record := range f.Records {
		if (!from.IsZero() && record.Timestamp.Before(from)) ||
			(!to.IsZero() && record.Timestamp.After(to)) {
			continue
		}
		filteredRecords = append(filteredRecords, record)
	}
	sort.Slice(filteredRecords, func(i, j int) bool {
		return filteredRecords[i].Timestamp.Before(filteredRecords[j].Timestamp)
	})
	return filteredRecords, nil
}

func (f *FakeStore) CountEntriesByTime(from time.Time, to time.Time) ([]db.TimestampCount, error) {
	count := make(map[time.Time]int)
	for _, record := range f.Records {
		if (!from.IsZero() && record.Timestamp.Before(from)) ||
			(!to.IsZero() && record.Timestamp.After(to)) {
			continue
		}
		count[record.Timestamp] = 1 + count[record.Timestamp]
	}
	timestampCount := make([]db.TimestampCount, 0, len(count))
	for timestamp, value := range count {
		timestampCount = append(timestampCount, db.TimestampCount{Timestamp: timestamp, Count: value})
	}
	sort.Slice(timestampCount, func(i, j int) bool {
		return timestampCount[i].Timestamp.Before(timestampCount[j].Timestamp)
	})
	return timestampCount, nil
}
