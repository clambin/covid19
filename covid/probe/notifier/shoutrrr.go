package notifier

import (
	"fmt"
	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/containrrr/shoutrrr/pkg/types"
	log "github.com/sirupsen/logrus"
)

// NotificationSender provides an interface to send notifications (e.g. to Shoutrrr)
//go:generate mockery --name NotificationSender
type NotificationSender interface {
	Send(title, message string) (err error)
}

// ShoutrrrSender implements the NotificationSender interface for Shoutrrr
type ShoutrrrSender struct {
	router *router.ServiceRouter
}

// NewNotificationSender creates a new ShoutrrrSender
func NewNotificationSender(url string) *ShoutrrrSender {
	sender, err := shoutrrr.CreateSender(url)
	if err != nil {
		log.WithError(err).Error("unable to create shoutrrr sender")
	}
	return &ShoutrrrSender{router: sender}
}

// Send a notification
func (s *ShoutrrrSender) Send(title, message string) (err error) {
	if s.router == nil {
		return fmt.Errorf("no sender")
	}
	params := types.Params{}
	params.SetTitle(title)
	errs := s.router.Send(message, &params)
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
