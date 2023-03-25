package shoutrrr_test

import (
	"github.com/clambin/covid19/covid/shoutrrr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestNotificationSender_Send(t *testing.T) {
	webhook := os.Getenv("SHOUTRRR_SLACK_URL")
	if webhook == "" {
		t.Skip("SHOUTRRR_SLACK_URL not set. Skipping test")
	}
	s, err := shoutrrr.NewRouter(webhook)
	require.NoError(t, err)
	err = s.Send("NotificationSender Test", "sent at "+time.Now().Format(time.RFC3339))
	assert.NoError(t, err)
}

func TestNotificationSender_Error(t *testing.T) {
	_, err := shoutrrr.NewRouter("invalid-url")
	assert.Error(t, err)
}
