package backfill

import (
	"encoding/json"
	"github.com/clambin/covid19/coviddb"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// Backfiller retrieves historic COVID19 data and adds it to the database
type Backfiller struct {
	Client *http.Client
	DB     coviddb.DB
}

// Create a new Backfiller object
func Create(db coviddb.DB) *Backfiller {
	return &Backfiller{Client: &http.Client{}, DB: db}
}

// Run the backfiller.  Get all supported countries from the API
// Then add any historical record that is older than the first
// record in the DB
func (backFiller *Backfiller) Run() error {
	var (
		err       error
		countries map[string]struct {
			Name string
			Code string
		}
		entries []struct {
			Confirmed int64
			Recovered int64
			Deaths    int64
			Date      time.Time
		}
		first time.Time
	)

	countries, err = backFiller.getCountries()
	if err == nil {
		var found bool
		if first, found, err = backFiller.DB.GetFirstEntry(); err == nil {
			log.Debugf("First entry in DB: %s", first.String())

			for slug, details := range countries {
				realName := lookupCountryName(details.Name)
				log.Debugf("Getting data for %s (slug: %s)", realName, slug)
				records := make([]coviddb.CountryEntry, 0)
				if entries, err = backFiller.getHistoricalData(slug); err == nil {
					for _, entry := range entries {
						log.Debugf("Entry date: %s", entry.Date.String())
						if !found || entry.Date.Before(first) {
							records = append(records, coviddb.CountryEntry{
								Timestamp: entry.Date,
								Code:      details.Code,
								Name:      realName,
								Confirmed: entry.Confirmed,
								Deaths:    entry.Deaths,
								Recovered: entry.Recovered})
						}
					}
					err = backFiller.DB.Add(records)
					if err == nil {
						log.Infof("Received data for %s. %d entries added", realName, len(records))
					}
				}
			}
		}
	}
	return err
}

const url = "https://api.covid19api.com"

func (backFiller *Backfiller) getCountries() (map[string]struct {
	Name string
	Code string
}, error) {
	var result = map[string]struct {
		Name string
		Code string
	}{}

	req, _ := http.NewRequest("GET", url+"/countries", nil)
	resp, err := backFiller.slowCall(req)

	if err == nil {
		var stats []struct {
			Country string
			Slug    string
			ISO2    string
		}

		decoder := json.NewDecoder(resp.Body)
		if err = decoder.Decode(&stats); err == nil {
			for _, entry := range stats {
				result[entry.Slug] = struct {
					Name string
					Code string
				}{Name: entry.Country, Code: entry.ISO2}
			}
		}
		_ = resp.Body.Close()
	}

	return result, err
}

func (backFiller *Backfiller) getHistoricalData(slug string) ([]struct {
	Confirmed int64
	Recovered int64
	Deaths    int64
	Date      time.Time
}, error) {
	var stats []struct {
		Confirmed int64
		Recovered int64
		Deaths    int64
		Date      time.Time
	}

	req, _ := http.NewRequest("GET", url+"/total/country/"+slug, nil)
	resp, err := backFiller.slowCall(req)

	if err == nil {
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&stats)
		_ = resp.Body.Close()
	}

	return stats, err
}

// slowCall handles 429 errors, slowing down before trying again
func (backFiller *Backfiller) slowCall(req *http.Request) (*http.Response, error) {
	resp, err := backFiller.Client.Do(req)

	for err == nil && resp.StatusCode == 429 {
		_ = resp.Body.Close()
		log.Debug("429 recv'd. Slowing down")
		time.Sleep(time.Second * 5)
		resp, err = backFiller.Client.Do(req)
	}

	if err == nil && resp.StatusCode == 200 {
		return resp, nil
	}

	return nil, err

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
