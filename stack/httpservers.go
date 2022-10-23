package stack

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

type HTTPServers map[string]*http.Server

func (h HTTPServers) Start(wg *sync.WaitGroup) {
	wg.Add(len(h))
	for name, server := range h {
		go func(name string, server *http.Server) {
			if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				log.WithError(err).Fatalf("unable to start %s HTTP server", name)
			}
			wg.Done()
		}(name, server)
	}
}

func (h HTTPServers) Stop(timeout time.Duration) {
	for name, server := range h {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		if err := server.Shutdown(ctx); err != nil {
			log.WithError(err).Errorf("unable to start %s HTTP server", name)
		}
		cancel()
	}
}
