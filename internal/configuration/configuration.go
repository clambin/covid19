package configuration

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type Configuration struct {
	Port     int                  `yaml:"port"`
	Debug    bool                 `yaml:"debug"`
	Postgres PostgresDB           `yaml:"postgres"`
	Monitor  MonitorConfiguration `yaml:"monitor"`
	Grafana  GrafanaConfiguration `yaml:"grafana"`
}

type PostgresDB struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type MonitorConfiguration struct {
	Enabled       bool                       `yaml:"enabled"`
	Interval      time.Duration              `yaml:"interval"`
	RapidAPIKey   ValueOrEnvVar              `yaml:"rapidAPIKey"`
	Notifications NotificationsConfiguration `yaml:"notifications"`
}

type ValueOrEnvVar struct {
	Value  string `yaml:"value"`
	EnvVar string `yaml:"envVar"`
}

type NotificationsConfiguration struct {
	Enabled   bool          `yaml:"enabled"`
	URL       ValueOrEnvVar `yaml:"url"`
	Countries []string      `yaml:"countries"`
}

type GrafanaConfiguration struct {
	Enabled bool `yaml:"enabled"`
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

// LoadConfigurationFile loads the configuration file from file
func LoadConfigurationFile(fileName string) (*Configuration, error) {
	var (
		err           error
		content       []byte
		configuration *Configuration
	)
	if content, err = ioutil.ReadFile(fileName); err == nil {
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
			Enabled:  true,
			Interval: 20 * time.Minute,
		},
	}
	err := yaml.Unmarshal(content, &configuration)

	if err == nil {
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
		pgUser = "covid"
	}
	if pgPassword = os.Getenv("pg_password"); pgPassword == "" {
		pgPassword = "covid"
	}

	return PostgresDB{
		Host:     pgHost,
		Port:     pgPort,
		Database: pgDatabase,
		User:     pgUser,
		Password: pgPassword,
	}
}
