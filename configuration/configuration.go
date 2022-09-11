package configuration

import (
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

// Configuration for covid19 app
type Configuration struct {
	Port     int                  `yaml:"port"`
	Debug    bool                 `yaml:"debug"`
	Postgres PostgresDB           `yaml:"postgres"`
	Monitor  MonitorConfiguration `yaml:"monitor"`
}

// PostgresDB configuration parameters
type PostgresDB struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
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
	RapidAPIKey   string                    `yaml:"rapidAPIKey"`
	Notifications NotificationConfiguration `yaml:"notifications"`
}

// NotificationConfiguration allows to set a notification when a country gets new data
type NotificationConfiguration struct {
	Enabled   bool     `yaml:"enabled"`
	URL       string   `yaml:"url"`
	Countries []string `yaml:"countries"`
}

// LoadConfiguration loads the configuration file from memory
func LoadConfiguration(content io.Reader) (*Configuration, error) {
	configuration := Configuration{
		Port: 8080,
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
