package covidprobe

import (
	"covid19/internal/coviddb"
)

type CovidProbe struct {
	db          coviddb.CovidDB
	apiKey      string
	pushGateway string
	measured    int
}

func NewCovidProbe(db coviddb.CovidDB, apiKey, pushGateway string) (*CovidProbe) {
	return &CovidProbe{db: db, apiKey: apiKey, pushGateway: pushGateway}
}

func (probe *CovidProbe) Run() (error) {
	err := probe.Measure()

	return err
}

func (probe *CovidProbe) Measure() (int error) {
	probe.measured = 1
	return nil
}

func (probe *CovidProbe) Measured() (int) {
	return probe.measured
}

func (probe *CovidProbe) Report() (error) {
	return nil
}

