package main

import (
	"errors"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/clambin/covid19/stack"
	"github.com/clambin/covid19/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slog"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	cmd, cfg, err := GetConfiguration("covid19", os.Args)
	if err != nil {
		panic(fmt.Errorf("failed to initialize application: %w", err))
	}

	var opts slog.HandlerOptions
	if cfg.Debug {
		opts.Level = slog.LevelDebug
		opts.AddSource = true
	}
	slog.SetDefault(slog.New(opts.NewTextHandler(os.Stdout)))

	slog.Info("covid19 starting", "version", version.BuildVersion)

	var s *stack.Stack
	if s, err = stack.CreateStack(cfg); err != nil {
		slog.Error("app init failed", "err", err)
		os.Exit(1)
	}
	prometheus.DefaultRegisterer.MustRegister(s)

	switch cmd {
	case handlerCmd.FullCommand():
		go runPrometheusServer(cfg.PrometheusPort)
		if err = s.RunHandler(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start simplejson handler", "err", err)
			os.Exit(1)
		}
	case loaderCmd.FullCommand():
		s.Load()
	case populationLoaderCmd.FullCommand():
		s.LoadPopulation()
	default:
		slog.Warn("invalid command", "command", cmd)
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
		return "", nil, fmt.Errorf("configuration: %w", err)
	}
	defer func() { _ = f.Close() }()

	if cfg, err = configuration.LoadConfiguration(f); err != nil {
		return "", nil, fmt.Errorf("load configuration: %w", err)
	}

	if debug {
		cfg.Debug = true
	}

	return
}

func runPrometheusServer(port int) {
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start Prometheus listener", "err", err)
	}
}
