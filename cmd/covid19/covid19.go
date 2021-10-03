package main

import (
	"context"
	"github.com/clambin/covid19/application"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/version"
	log "github.com/sirupsen/logrus"
	"github.com/xonvanetta/shutdown/pkg/shutdown"
	"gopkg.in/alecthomas/kingpin.v2"
	_ "net/http/pprof"
	"os"
	"path/filepath"
)

func main() {
	var (
		cfg            *configuration.Configuration
		debug          bool
		configFileName string
	)

	log.WithField("version", version.BuildVersion).Info("covid19 monitor starting")
	a := kingpin.New(filepath.Base(os.Args[0]), "covid19 monitor")

	a.Version(version.BuildVersion)
	a.HelpFlag.Short('h')
	a.VersionFlag.Short('v')
	a.Flag("debug", "Log debug messages").BoolVar(&debug)
	a.Flag("config", "Configuration file").Required().StringVar(&configFileName)

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	if cfg, err = configuration.LoadConfigurationFile(configFileName); err != nil {
		log.WithField("err", err).Fatal("Failed to read config file")
	}

	if debug || cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	stack := application.New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go stack.Run(ctx)
	<-shutdown.Chan()
}
