package main

import (
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/stack"
	"github.com/clambin/covid19/version"
	log "github.com/sirupsen/logrus"
	"github.com/xonvanetta/shutdown/pkg/shutdown"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
)

func main() {
	cmd, cfg, err := GetConfiguration("covid19", os.Args)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize application")
	}

	var s *stack.Stack
	if s, err = stack.CreateStack(cfg); err != nil {
		log.WithError(err).Fatal("app init failed")
	}

	switch cmd {
	case handlerCmd.FullCommand():
		go s.RunHandler()
		<-shutdown.Chan()
		s.StopHandler()
	case loaderCmd.FullCommand():
		s.Load()
	case populationLoaderCmd.FullCommand():
		s.LoadPopulation()
	default:
		log.Fatalf("invalid command: %s", cmd)
	}
}

var (
	handlerCmd          *kingpin.CmdClause
	loaderCmd           *kingpin.CmdClause
	populationLoaderCmd *kingpin.CmdClause
)

// GetConfiguration parses the provided commandline arguments and creates the required configuration
func GetConfiguration(application string, args []string) (cmd string, cfg *configuration.Configuration, err error) {
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
	a.Flag("config", "Configuration file").Required().ExistingFileVar(&configFileName)
	handlerCmd = a.Command("handler", "runs the simplejson handler")
	loaderCmd = a.Command("loader", "retrieves new covid data")
	populationLoaderCmd = a.Command("population", "retrieves latest population data")

	cmd, err = a.Parse(args[1:])
	if err != nil {
		a.Usage(args[1:])
	}

	var f *os.File
	if f, err = os.OpenFile(configFileName, os.O_RDONLY, 0); err != nil {
		log.WithField("err", err).Fatal("Failed to access config file")
	}
	defer func() { _ = f.Close() }()

	if cfg, err = configuration.LoadConfiguration(f); err != nil {
		log.WithField("err", err).Fatal("Invalid config file")
	}

	if debug {
		cfg.Debug = true
	}

	return
}
