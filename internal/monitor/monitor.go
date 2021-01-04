package monitor

import (
	log "github.com/sirupsen/logrus"

	"time"

	"covid19/internal/covidprobe"
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
	Reports          struct {
		Countries struct {
			PushGateway string
		}
		Updates struct {
			Token     string
			User      string
			Countries []string
		}
	}
	ProfileName string
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
