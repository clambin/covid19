package configuration

import (
	"github.com/clambin/covid19/version"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
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

// GetConfiguration parses the provided commandline arguments and creates the required configuration
func GetConfiguration(application string, args []string) (cfg *Configuration) {
	var (
		debug          bool
		configFileName string
	)

	log.WithField("version", version.BuildVersion).Info(application + " starting")
	a := kingpin.New(filepath.Base(args[0]), application)

	a.Version(version.BuildVersion)
	a.HelpFlag.Short('h')
	a.VersionFlag.Short('v')
	a.Flag("debug", "Log debug messages").BoolVar(&debug)
	a.Flag("config", "Configuration file").Required().StringVar(&configFileName)

	_, err := a.Parse(args[1:])
	if err != nil {
		a.Usage(args[1:])
		os.Exit(1)
	}

	if cfg, err = LoadConfigurationFile(configFileName); err != nil {
		log.WithField("err", err).Fatal("Failed to read config file")
	}

	if debug {
		cfg.Debug = true
	}

	return
}
