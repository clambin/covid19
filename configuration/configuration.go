package configuration

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"time"
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

// MonitorConfiguration parameters
type MonitorConfiguration struct {
	Interval      time.Duration             `yaml:"interval"`
	RapidAPIKey   ValueOrEnvVar             `yaml:"rapidAPIKey"`
	Notifications NotificationConfiguration `yaml:"notifications"`
}

// ValueOrEnvVar allows a value to be specified directly, or via environment variable
type ValueOrEnvVar struct {
	Value  string `yaml:"value"`
	EnvVar string `yaml:"envVar"`
}

// NotificationConfiguration allows to set a notification when a country gets new data
type NotificationConfiguration struct {
	Enabled   bool          `yaml:"enabled"`
	URL       ValueOrEnvVar `yaml:"url"`
	Countries []string      `yaml:"countries"`
}

// Set a ValueOrEnvVar
func (v *ValueOrEnvVar) Set() {
	if v.EnvVar != "" {
		v.Value = os.Getenv(v.EnvVar)

		if v.Value == "" {
			log.WithField("envVar", v.EnvVar).Warning("environment variable not set")
		}
	}
}

// Get a ValueOrEnvVar
func (v ValueOrEnvVar) Get() (value string) {
	value = v.Value
	if v.EnvVar != "" {
		value = os.Getenv(v.EnvVar)
	}
	return value
}

// LoadConfigurationFile loads the configuration file from file
func LoadConfigurationFile(fileName string) (configuration *Configuration, err error) {
	var content []byte
	if content, err = os.ReadFile(fileName); err == nil {
		configuration, err = LoadConfiguration(content)
	}
	return configuration, err
}

// LoadConfiguration loads the configuration file from memory
func LoadConfiguration(content []byte) (*Configuration, error) {
	configuration := Configuration{
		Port:     8080,
		Postgres: loadPGEnvironment(),
		Monitor: MonitorConfiguration{
			Interval: 20 * time.Minute,
		},
	}
	err := yaml.Unmarshal(content, &configuration)

	if err == nil {
		// make postgres password a ValueOrEnvVar too
		configuration.Monitor.RapidAPIKey.Set()
		configuration.Monitor.Notifications.URL.Set()
	}

	log.WithField("err", err).Debug("LoadConfiguration")

	return &configuration, err
}

// loadPGEnvironment loads Postgres config from environment variables
func loadPGEnvironment() PostgresDB {
	var (
		err        error
		pgHost     string
		pgPort     int
		pgDatabase string
		pgUser     string
		pgPassword string
	)
	if pgHost = os.Getenv("pg_host"); pgHost == "" {
		pgHost = "postgres"
	}
	if pgPort, err = strconv.Atoi(os.Getenv("pg_port")); err != nil || pgPort == 0 {
		pgPort = 5432
	}
	if pgDatabase = os.Getenv("pg_database"); pgDatabase == "" {
		pgDatabase = "covid19"
	}
	if pgUser = os.Getenv("pg_user"); pgUser == "" {
		pgUser = "probe"
	}
	if pgPassword = os.Getenv("pg_password"); pgPassword == "" {
		pgPassword = "probe"
	}

	return PostgresDB{
		Host:     pgHost,
		Port:     pgPort,
		Database: pgDatabase,
		User:     pgUser,
		Password: pgPassword,
	}
}
