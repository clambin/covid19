package backfill

import (
	"encoding/json"
	"fmt"
	"github.com/clambin/covid19/db"
	"github.com/clambin/covid19/models"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// Backfiller retrieves historic COVID19 data and adds it to the database
type Backfiller struct {
	URL   string
	Store db.CovidStore
}

// New creates a new Backfiller object
func New(store db.CovidStore) *Backfiller {
	return &Backfiller{URL: covid19url, Store: store}
}

// Run the backfiller.  Get all supported countries from the API
// Then add any historical record that is older than the first
// record in the CovidDB
func (backFiller *Backfiller) Run() error {

	countries, err := backFiller.getCountries()
	if err != nil {
		return fmt.Errorf("could not retrieve supported countries: %s", err.Error())
	}

	for slug, details := range countries {
		records := make([]models.CountryEntry, 0)
		realName := lookupCountryName(details.Name)
		log.Debugf("Getting data for %s (slug: %s)", realName, slug)

		var entries []struct {
			Confirmed int64
			Recovered int64
			Deaths    int64
			Date      time.Time
		}
		entries, err = backFiller.getHistoricalData(slug)

		if err != nil {
			log.WithError(err).WithField("country", slug).Warning("failed to get history")
			continue
		}

		for _, entry := range entries {
			records = append(records, models.CountryEntry{
				Timestamp: entry.Date.Add(24 * time.Hour),
				Code:      details.Code,
				Name:      realName,
				Confirmed: entry.Confirmed,
				Deaths:    entry.Deaths,
				Recovered: entry.Recovered})
		}

		err = backFiller.Store.Add(records)
		if err == nil {
			log.Infof("Received data for %s. %d entries added", realName, len(records))
		}
	}
	return err
}

const covid19url = "https://api.covid19api.com"

func (backFiller *Backfiller) getCountries() (result map[string]struct{ Name, Code string }, err error) {
	req, _ := http.NewRequest(http.MethodGet, backFiller.URL+"/countries", nil)
	var resp *http.Response
	resp, err = backFiller.slowCall(req)

	if err == nil {
		if resp.StatusCode == http.StatusOK {
			var stats []struct {
				Country string
				Slug    string
				ISO2    string
			}

			decoder := json.NewDecoder(resp.Body)
			if err = decoder.Decode(&stats); err == nil {
				result = make(map[string]struct{ Name, Code string })
				for _, entry := range stats {
					result[entry.Slug] = struct {
						Name string
						Code string
					}{Name: entry.Country, Code: entry.ISO2}
				}
			}
		}
		_ = resp.Body.Close()
	}

	return result, err
}

func (backFiller *Backfiller) getHistoricalData(slug string) (stats []struct {
	Confirmed int64
	Recovered int64
	Deaths    int64
	Date      time.Time
}, err error) {
	req, _ := http.NewRequest(http.MethodGet, backFiller.URL+"/total/country/"+slug, nil)
	var resp *http.Response
	resp, err = backFiller.slowCall(req)

	if err == nil {
		if resp.StatusCode == http.StatusOK {
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&stats)
		}
		_ = resp.Body.Close()
	}

	return stats, err
}

// slowCall handles 429 errors, slowing down before trying again
func (backFiller *Backfiller) slowCall(req *http.Request) (resp *http.Response, err error) {
	client := &http.Client{}
	resp, err = client.Do(req)

	waitTime := 250 * time.Millisecond

	for err == nil && resp.StatusCode == http.StatusTooManyRequests {
		_ = resp.Body.Close()

		log.WithField("waitTime", waitTime).Debug("429 recv'd. Slowing down")
		time.Sleep(waitTime)

		resp, err = client.Do(req)

		if waitTime < 5*time.Second {
			waitTime *= 2
		}
	}

	return
}

// rapidapi's Covid API uses different country names than covidapi
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
