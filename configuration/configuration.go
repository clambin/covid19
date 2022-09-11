package configuration

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

// Configuration for covid19 app
type Configuration struct {
	Port     int                  `yaml:"port"`
	Debug    bool                 `yaml:"debug"`
	Postgres PostgresDB           `yaml:"postgres"`
	Monitor  MonitorConfiguration `yaml:"monitor"`
}

// MonitorConfiguration parameters
type MonitorConfiguration struct {
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
		Postgres: LoadPGEnvironmentWithDefaults(),
		Monitor:  MonitorConfiguration{},
	}
	err := yaml.Unmarshal(content, &configuration)

	if err == nil {
		// TODO: make postgres password a ValueOrEnvVar too
		configuration.Monitor.RapidAPIKey.Set()
		configuration.Monitor.Notifications.URL.Set()
	}

	log.WithField("err", err).Debug("LoadConfiguration")

	return &configuration, err
}
