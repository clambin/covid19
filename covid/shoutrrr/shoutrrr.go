package shoutrrr

import (
	"fmt"
	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/containrrr/shoutrrr/pkg/types"
	"strings"
)

// Sender interface for underlying notification routers
//
//go:generate mockery --name Sender
type Sender interface {
	Send(title, message string) (err error)
}

// Router implements the Sender interface for Shoutrrr
type Router struct {
	router *router.ServiceRouter
}

// NewRouter creates a new Router
func NewRouter(url string) (*Router, error) {
	r, err := shoutrrr.CreateSender(url)
	if err != nil {
		return nil, fmt.Errorf("shoutrrr: %w", err)
	}
	return &Router{router: r}, nil
}

// Send a notification
func (s *Router) Send(title, message string) error {
	if s.router == nil {
		return fmt.Errorf("router not initialized")
	}

	params := types.Params{}
	params.SetTitle(title)
	errs := s.router.Send(message, &params)
	var errors []string
	for _, err := range errs {
		if err != nil {
			errors = append(errors, err.Error())
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("router: send failed: %s", strings.Join(errors, ","))
	}
	return nil
}
