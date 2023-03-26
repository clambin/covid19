package configuration

import (
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

// Configuration for covid19 app
type Configuration struct {
	Postgres       PostgresDB           `yaml:"postgres"`
	Monitor        MonitorConfiguration `yaml:"monitor"`
	Port           int                  `yaml:"port"`
	PrometheusPort int                  `yaml:"prometheusPort"`
	Debug          bool                 `yaml:"debug"`
}

// PostgresDB configuration parameters
type PostgresDB struct {
	Host     string `yaml:"host"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
}

// IsValid checks if the postgres configuration is valid
func (pg PostgresDB) IsValid() bool {
	return pg.Host != "" &&
		pg.Port != 0 &&
		pg.Database != "" &&
		pg.User != "" &&
		pg.Password != ""
}

// MonitorConfiguration parameters
type MonitorConfiguration struct {
	Notifications NotificationConfiguration `yaml:"notifications"`
	RapidAPIKey   string                    `yaml:"rapidAPIKey"`
}

// NotificationConfiguration allows to set a notification when a country gets new data
type NotificationConfiguration struct {
	Countries []string `yaml:"countries"`
	URL       string   `yaml:"url"`
	Enabled   bool     `yaml:"enabled"`
}

// LoadConfiguration loads the configuration file from memory
func LoadConfiguration(content io.Reader) (*Configuration, error) {
	configuration := Configuration{
		Port:           8080,
		PrometheusPort: 9090,
		Postgres: PostgresDB{
			Host:     "postgres",
			Port:     5432,
			Database: "covid19",
			User:     "covid",
		},
		Monitor: MonitorConfiguration{},
	}
	body, err := io.ReadAll(content)
	if err == nil {
		body = []byte(os.ExpandEnv(string(body)))
		err = yaml.Unmarshal(body, &configuration)
	}

	return &configuration, err
}
