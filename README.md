# covid19
![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/clambin/covid19?color=green&label=Release&style=plastic)
![Build)](https://github.com/clambin/covid19/workflows/Build/badge.svg)
![Codecov](https://img.shields.io/codecov/c/gh/clambin/covid19?style=plastic)
![Go Report Card](https://goreportcard.com/badge/github.com/clambin/covid19)
![GitHub](https://img.shields.io/github/license/clambin/covid19?style=plastic)


A lightweight Covid-19 tracker.

## Introduction
This package tracks global Covid-19 data. The basic operation is as follows:

- It retrieves daily updates from a public Covid-19 tracker and stores them in an external Postgres DB
- Grafana is used to visualize the data

## Installation
A Docker image for the main Covid-19 tracker is available on [docker](https://hub.docker.com/r/clambin/covid19). Images are available for amd64, arm64 & arm32.

Binaries are also available at [github](https://github.com/clambin/covid19/releases).

Alternatively, you can clone the repository and build from source:

```
git clone https://github.com/clambin/covid19.git
cd covid19
go build cmd/covid19/covid19.go
go build cmd/backfill/backfill.go
```

You will need to have Go 1.16 installed on your system.

## Configuration
### Configuration file

Use the following yaml file to configure parameters & desired behaviour:

```
# HTTP port for Prometheus (optional) & Grafana. Default is 8080
port: 8080
# Turn on debug logging. Default is false
debug: false
# Configuration for Postgres DB
postgres:
  # Postgres Host IP or host
  host: postgres
  # Postgres Port. Default is 5432
  port: 5432
  # Postgres database name. Default is "covid19"
  database: "covid19"
  # Postgres owner of the database. Default is "covid"
  user: "covid"
  # Password for user. 
  # Alternatively, can be provided through pg_password environment variable
  password: "its4covid"
# Monitor section to configure how new covid data should be retrieved
monitor:
  # Turn on covid data capturing. Default is true
  enabled: true
  # How ofter new data should be gathered. Default is 20m.
  # Be kind to API providers and don't put this too low. Data only changes daily anyway
  interval: 20m
  # API Key for the APIs. See below.
  rapidAPIKey:
    value: "long-rapid-api-key"
    # alternatively, specify an environment variable that holds the API Key:
    # envVar: "KEY_ENV_VAR_NAME"
  # covid19 can be configured to send a notification when new data is found for a set of countries
  notifications:
    # Turn on notifications. Default is true
    enabled: true
    # URL to send notifications to. See https://github.com/containrrr/shoutrrr for options
    url:
      value: https://hooks.slack.com/services/token1/token2/token3
      # alternatively, specify an environment variable that holds the API Key:
      # envVar: "URL_ENV_VAR_NAME"
    # List of country names for which to send an event
    countries:
      - Belgium
      - US
```

## Postgres
Covid19 uses a Postgres database to store collected data. Create a database and postgres user with permissions to create new tables & indexes. 
Covid19 will handle table creation itself. 

## RapidAPI
Covid19 uses two APIs published on RapidAPI.com to collect new data. You will need to create an account, which will give you an API Key. 
Add this key to the configuration file above and subscribe to the following two services:

- https://rapidapi.com/KishCom/api/covid-19-coronavirus-statistics
- https://rapidapi.com/aldair.sr99/api/world-population/

The first one offers the latest Covid-19 statistics. The second one provides population figures for each country.

## Grafana data sources
Both Postgres and the covid19 Grafana data source will need to be configured in Grafana. 
This can be done manually through the Grafana admin UI, or through a datasource provisioning file, e.g.

```
apiVersion: 1
datasources:
  - id: 5
    orgid: 1
    name: covid19api
    type: grafana-simple-json-datasource
    access: proxy
    # URL of covid19 server
    url: http://covid19.default.svc:5000
    password: ""
    user: ""
    database: ""
    basicauth: false
    basicauthuser: null
    basicauthpassword: null
    isdefault: false
    jsondata: {}
    securejsondata: null
  - id: 3
    orgId: 1
    name: PostgreSQL
    type: postgres
    # URL of Postgres DB server
    url: postgres.default:5432
    # Database name, as defined above
    database: covid19
    # Database userm as defined above
    user: grafana
    secureJsonData:
      # Database user password
      password: "your-password-here"
    jsonData:
      sslmode: "disable"
```

### Backfilling historical data
Covid19 only adds new data for the current day. For new installations, historical data will need to be added manually. 
Use the provided utility to do this:

```
backfill --postgres-host=<postgres-host> \
         --postgres-port=<postgres-port> \
         --postgres-user=<postgres-user>
         --postgres-password=<postgres-password>
```

When sticking to the default port & user, those arguments can be omitted.

## Running covid19
### Command-line options
The following command-line arguments can be passed:

```
usage: covid19 --config=CONFIG [<flags>]

covid19

Flags:
-h, --help           Show context-sensitive help (also try --help-long and --help-man).
-v, --version        Show application version.
--debug          Log debug messages
--config=CONFIG  Configuration file
```

## Grafana
The repo contains two sample [dashboards](assets/grafana/dashboards). One dashboard provides a view per country.
The second one provides an overview of cases, evolution, per capita stats across the world.

Feel free to customize as you see fit.

## Authors

- Christophe Lambin

## License

This project is licensed under the MIT License - see the [license](LICENSE.md) file for details.
