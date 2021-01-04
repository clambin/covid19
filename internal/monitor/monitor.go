package monitor

import (
	"covid19/internal/covidprobe"
	popprobe "covid19/internal/population/probe"
	"covid19/internal/reporters"

	log "github.com/sirupsen/logrus"

	"time"
)

type Configuration struct {
	Debug       bool
	Once        bool
	Interval    time.Duration
	PushGateway string

	ProfileName string
	Postgres    struct {
		Host     string
		Port     int
		Database string
		User     string
		Password string
	}
	RapidAPI struct {
		Key string
	}
	Reports reporters.ReportsConfiguration
}

func Run(cfg *Configuration, covidProbe *covidprobe.Probe, popProbe *popprobe.Probe) bool {
	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	covidDone := make(chan bool)
	go func() {
		err := covidProbe.Run()
		if err != nil {
			log.Warningf("covid probe error: %s", err)
		}
		covidDone <- err == nil
	}()

	popDone := make(chan bool)
	go func() {
		err := popProbe.Run()
		if err != nil {
			log.Warningf("population probe error: %s", err)
		}
		popDone <- err == nil
	}()

	popOK := <-popDone
	covidOK := <-covidDone

	return popOK && covidOK
}
