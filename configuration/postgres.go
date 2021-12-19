package configuration

import (
	"os"
	"strconv"
)

// PostgresDB configuration parameters
type PostgresDB struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func LoadPGEnvironmentWithDefaults() (config PostgresDB) {
	config = LoadPGEnvironment()
	if config.Host == "" {
		config.Host = "postgres"
	}
	if config.Port == 0 {
		config.Port = 5432
	}
	if config.Database == "" {
		config.Database = "covid19"
	}
	if config.User == "" {
		config.User = "covid"
	}
	if config.Password == "" {
		config.Password = "covid"
	}
	return
}

// LoadPGEnvironment loads Postgres config from environment variables
func LoadPGEnvironment() PostgresDB {
	var envNames = []string{
		"pg_host",
		"pg_port",
		"pg_database",
		"pg_user",
		"pg_password",
	}

	envValues := make(map[string]string)
	for _, envName := range envNames {
		envValues[envName] = os.Getenv(envName)
	}
	port, _ := strconv.Atoi(envValues["pg_port"])
	return PostgresDB{
		Host:     envValues["pg_host"],
		Port:     port,
		Database: envValues["pg_database"],
		User:     envValues["pg_user"],
		Password: envValues["pg_password"],
	}
}

func (pg PostgresDB) IsValid() bool {
	return pg.Host != "" &&
		pg.Port != 0 &&
		pg.Database != "" &&
		pg.User != "" &&
		pg.Password != ""
}
