# covid19
![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/clambin/covid19?color=green&label=Release&style=plastic)
![Build)](https://github.com/clambin/covid19/workflows/Build/badge.svg)
![Codecov](https://img.shields.io/codecov/c/gh/clambin/covid19?style=plastic)
![Go Report Card](https://goreportcard.com/badge/github.com/clambin/covid19)
![GitHub](https://img.shields.io/github/license/clambin/covid19?style=plastic)


A lightweight Covid-19 tracker.

## Introduction
This package tracks global Covid-19 data. It provides three commands:

- loader: retrieves the latest updates from a public Covid-19 tracker and stores them in an external Postgres DB
- population: retrieves the latest population figures from a public tracker and stores them in an external Postgres DB
- handler: implements the targets for the provided Grafana dashboards

## Installation
Docker images are available on ghcr.io:

- [covid19](https://github.com/clambin/covid19/pkgs/container/covid19)

Images are available for amd64, arm64 & arm32. Binaries are also available on [github](https://github.com/clambin/covid19/releases).

A helm chart is available at https://clambin.github.io/helm-charts. It runs the handler as a deployment and configures
two CronJobs to load covid and population data on a daily basis.

## Configuration
### Configuration file

Use the following yaml file to configure parameters & desired behaviour:

```
# HTTP port for Grafana SimpleJSON server. Default is 8080
port: 8080
# HTTP port for Prometheus metrics. Default is 9090.
prometheusPort: 9090
# Turn on debug logging. Default is false
debug: false
# Configuration for Postgres DB
postgres:
  # Postgres Host IP or host. Default is 'postgres'
  host: postgres
  # Postgres Port. Default is 5432
  port: 5432
  # Postgres database name. Default is "covid19"
  database: "covid19"
  # Postgres owner of the database. Default is "covid"
  user: "covid"
  # Password for user. 
  password: "some-password"
# Monitor section to configure how new covid data should be retrieved
monitor:
  # API Key for the APIs. See below.
  rapidAPIKey: "rapid-api-key"
  # covid19 can be configured to send a notification when new data is found for a set of countries
  notifications:
    # Turn on notifications. Default is false
    enabled: true
    # URL to send notifications to. See https://github.com/containrrr/shoutrrr for options
    url: https://hooks.slack.com/services/token1/token2/token3
    # List of country names for which to send an event
    countries:
      - Belgium
      - US
```

covid19 will substitute any environment variables referenced in the configuration file. E.g.:

```
postgres:
  host: postgres
  port: 5432
  database: "covid19"
  user: "covid"
  password: "$pg_password"
```

will use the value of the environment variable 'pg_password' is the password for the Postgres DB.

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
The covid19 Grafana data source will need to be configured in Grafana. This can be done manually through the Grafana admin UI, or through a datasource provisioning file, e.g.

```
apiVersion: 1
datasources:
  - id: 5
    orgid: 1
    name: covid19
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
```

## Running
### Command-line options
Each mode supports the same command-line arguments:

```
usage: covid19 --config=CONFIG [<flags>] <command> [<args> ...]

covid19

Flags:
  -h, --help           Show context-sensitive help (also try --help-long and --help-man).
  -v, --version        Show application version.
      --debug          Log debug messages
      --config=CONFIG  Configuration file

Commands:
  help [<command>...]
    Show help.

  handler
    runs the simplejson handler

  loader
    retrieves new covid data

  population
    retrieves latest population data
```

## Grafana
The repo contains sample [dashboards](assets/grafana/dashboards). One dashboard provides a view per country.
A second one provides an overview of cases, evolution, per capita stats across the world.

## Authors

- Christophe Lambin

## License

This project is licensed under the MIT License - see the [license](LICENSE.md) file for details.
