package monitor

import (
	log "github.com/sirupsen/logrus"

	"time"

	covidprobe "covid19/internal/covid/probe"
	popprobe "covid19/internal/population/probe"
)

type Configuration struct {
	Debug            bool
	Once             bool
	Interval         time.Duration
	PostgresHost     string
	PostgresPort     int
	PostgresDatabase string
	PostgresUser     string
	PostgresPassword string
	APIKey           string
	PushGateway      string
	ProfileName      string
}

func Run(cfg *Configuration, covidProbe *covidprobe.Probe, popProbe *popprobe.Probe) {
	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	covidDone := make(chan bool)
	go func() {
		for {
			if err := covidProbe.Run(); err != nil {
				log.Warningf("covid probe error: %s", err)
			}
			if cfg.Once {
				covidDone <- true
			} else {
				time.Sleep(cfg.Interval)
			}
		}
	}()

	popDone := make(chan bool)
	go func() {
		for {
			if err := popProbe.Run(); err != nil {
				log.Warningf("population probe error: %s", err)
			}
			if cfg.Once {
				popDone <- true
			} else {
				time.Sleep(cfg.Interval)
			}
		}
	}()

	<-popDone
	<-covidDone
}
