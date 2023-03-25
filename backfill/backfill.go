package backfill

import (
	"github.com/clambin/covid19/models"
	"golang.org/x/exp/slog"
	"time"
)

// Backfiller retrieves historic COVID19 data and adds it to the database
type Backfiller struct {
	Client CovidGetter
	Store  CovidStoreAdder
}

type CovidStoreAdder interface {
	Add([]models.CountryEntry) error
}

type CovidGetter interface {
	GetCountries() (Countries, error)
	GetHistoricalData(string) ([]CountryData, error)
}

// New creates a new Backfiller object
func New(store CovidStoreAdder) *Backfiller {
	return &Backfiller{
		Client: Client{URL: covid19url},
		Store:  store}
}

// Run the backfiller.  Get all supported countries from the API
// Then add any historical record that is older than the first
// record in the DB
func (b *Backfiller) Run() error {

	countries, err := b.Client.GetCountries()
	if err != nil {
		return err
	}

	for slug, details := range countries {
		realName := lookupCountryName(details.Name)
		slog.Debug("Getting country data", "name", realName, "slug", slug)

		var entries []CountryData
		if entries, err = b.Client.GetHistoricalData(slug); err != nil {
			slog.Error("failed to get history", "err", err, "country", slug)
			continue
		}

		var records []models.CountryEntry
		for _, entry := range entries {
			records = append(records, models.CountryEntry{
				Timestamp: entry.Date.Add(24 * time.Hour),
				Code:      details.Code,
				Name:      realName,
				Confirmed: entry.Confirmed,
				Deaths:    entry.Deaths,
				Recovered: entry.Recovered})
		}

		if err = b.Store.Add(records); err == nil {
			slog.Info("Received country data ", "name", realName, "count", len(records))
		}
	}
	return err
}

const covid19url = "https://api.covid19api.com"

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
