package backfill

import (
	"covid19/internal/covid/db"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// Backfiller retrieves historic COVID19 data and adds it to the database
type Backfiller struct {
	client *http.Client
	db     db.DB
}

// Create a new Backfiller object
func Create(db db.DB) *Backfiller {
	return CreateWithClient(db, &http.Client{})
}

// CreateWithClient creates  a new Backfiller object w/ a supplier http.Client.
// Used for unit testtools
func CreateWithClient(db db.DB, client *http.Client) *Backfiller {
	return &Backfiller{client: client, db: db}
}

// Run the backfiller.  Get all supported countries from the API
// Then add any historical record that is older than the first
// record in the DB
func (backfiller *Backfiller) Run() error {
	countries, err := backfiller.getCountries()
	if err != nil {
		return err
	}

	first, err := backfiller.db.GetFirstEntry()
	if err != nil {
		return err
	}

	log.Debugf("First entry in DB: %s", first.String())

	for slug, details := range countries {
		realName := lookupCountryName(details.Name)

		log.Debugf("Getting data for %s (slug: %s)", realName, slug)

		records := make([]db.CountryEntry, 0)
		entries, err := backfiller.getHistoricalData(slug)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			log.Debugf("Entry date: %s", entry.Date.String())
			if entry.Date.Before(first) {
				records = append(records, db.CountryEntry{
					Timestamp: entry.Date,
					Code:      details.Code,
					Name:      realName,
					Confirmed: entry.Confirmed,
					Deaths:    entry.Deaths,
					Recovered: entry.Recovered})
			}
		}
		err = backfiller.db.Add(records)
		if err != nil {
			return err
		}
		log.Infof("Received data for %s. %d entries added", realName, len(records))
	}

	return err
}

const url = "https://api.covid19api.com"

func (backfiller *Backfiller) getCountries() (map[string]struct {
	Name string
	Code string
}, error) {
	req, _ := http.NewRequest("GET", url+"/countries", nil)

	for {
		resp, err := backfiller.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			log.Debug("429 recv'd. Slowing down")
			time.Sleep(time.Second * 2)
			continue
		}

		if resp.StatusCode != 200 {
			return nil, errors.New(resp.Status)
		}

		var stats []struct {
			Country string
			Slug    string
			ISO2    string
		}

		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&stats)

		if err != nil {
			return nil, err
		}

		result := make(map[string]struct {
			Name string
			Code string
		}, 0)
		for _, entry := range stats {
			result[entry.Slug] = struct {
				Name string
				Code string
			}{Name: entry.Country, Code: entry.ISO2}
		}
		return result, nil

	}
}

func (backfiller *Backfiller) getHistoricalData(slug string) ([]struct {
	Confirmed int64
	Recovered int64
	Deaths    int64
	Date      time.Time
}, error) {
	req, _ := http.NewRequest("GET", url+"/total/country/"+slug, nil)

	for {
		resp, err := backfiller.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			log.Debug("429 recv'd. Slowing down")
			time.Sleep(time.Second * 5)
			continue
		}

		if resp.StatusCode != 200 {
			return nil, errors.New(resp.Status)
		}

		var stats []struct {
			Confirmed int64
			Recovered int64
			Deaths    int64
			Date      time.Time
		}

		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&stats)

		return stats, nil

	}
}

// rapidapi's Covid API uses different names than covidapi
var (
	lookupTable = map[string]string{
		"Wallis and Futuna Islands":       "Wallis and Futuna",
		"Republic of Kosovo":              "Kosovo",
		"United States of America":        "US",
		"Holy See (Vatican City State)":   "Holy See",
		"Korea (South)":                   "Korea, South",
		"Saint-Martin (French part)":      "Saint Martin",
		"Cocos (Keeling) Islands":         "Cocos [Keeling] Islands",
		"Côte d'Ivoire":                   "Cote d'Ivoire",
		"Micronesia, Federated States of": "Micronesia",
		"Palestinian Territory":           "West Bank and Gaza",
		"Russian Federation":              "Russia",
		"Macao, SAR China":                "Macau",
		"ALA Aland Islands":               "Åland",
		"Pitcairn":                        "Pitcairn Islands",
		"Brunei Darussalam":               "Brunei",
		"Hong Kong, SAR China":            "Hong Kong",
		"Macedonia, Republic of":          "North Macedonia",
		"Virgin Islands, US":              "U.S. Virgin Islands",
		"Myanmar":                         "Burma",
		"Korea (North)":                   "North Korea",
		"Saint Vincent and Grenadines":    "Saint Vincent and the Grenadines",
		"Heard and Mcdonald Islands":      "Heard Island and McDonald Islands",
		"Svalbard and Jan Mayen Islands":  "Svalbard and Jan Mayen",
		"Taiwan, Republic of China":       "Taiwan*",
		"Tanzania, United Republic of":    "Tanzania",
		"Syrian Arab Republic (Syria)":    "Syria",
		"Iran, Islamic Republic of":       "Iran",
		"Venezuela (Bolivarian Republic)": "Venezuela",
		"Viet Nam":                        "Vietnam",
		"Falkland Islands (Malvinas)":     "Falkland Islands [Islas Malvinas]",
		"US Minor Outlying Islands":       "U.S. Minor Outlying Islands",
		"Lao PDR":                         "Laos",
		"Czech Republic":                  "Czechia",
		"Cape Verde":                      "Cabo Verde",
		"Swaziland":                       "Eswatini",
	}
)

func lookupCountryName(name string) string {
	converted, ok := lookupTable[name]
	if ok {
		return converted
	}
	return name
}
