package configuration

import (
	log "github.com/sirupsen/logrus"
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

// UnmarshalYAML unmarshalls a ValueOrEnvVar.  If the value is just a string, it will be converted to a ValueOrEnvVar.
func (v *ValueOrEnvVar) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		v.Value = value.Value
		v.EnvVar = ""
		return nil
	}
	type tmp ValueOrEnvVar
	var v2 tmp
	if err := value.Decode(&v2); err != nil {
		return err
	}
	*v = ValueOrEnvVar(v2)
	return nil
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
func (v *ValueOrEnvVar) Get() (value string) {
	value = v.Value
	if v.EnvVar != "" {
		value = os.Getenv(v.EnvVar)
	}
	return value
}

// LoadConfiguration loads the configuration file from memory
func LoadConfiguration(content io.Reader) (*Configuration, error) {
	configuration := Configuration{
		Port:     8080,
		Postgres: LoadPGEnvironmentWithDefaults(),
		Monitor:  MonitorConfiguration{},
	}
	body, err := io.ReadAll(content)
	if err != nil {
		return nil, err
	}

	body = []byte(os.ExpandEnv(string(body)))

	if err = yaml.Unmarshal(body, &configuration); err == nil {
		configuration.Monitor.RapidAPIKey.Set()
		configuration.Monitor.Notifications.URL.Set()
	}

	return &configuration, err
}
