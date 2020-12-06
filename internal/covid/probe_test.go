package covid_test

import (
	"testing"
	"github.com/stretchr/testify/assert"

	"covid19/internal/covid"
	"covid19/internal/covid/mock"
)

func TestProbe(t *testing.T) {
	apiClient := makeClient()
	db        := mock.Create(testDBData)
	probe     := covid.NewProbe(apiClient, db, nil)

	err := probe.Run()

	assert.Equal(t, nil, err)
	latest, err := db.ListLatestByCountry()
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(latest))
	assert.Equal(t, true, latest["A"].Equal(lastUpdate))
	assert.Equal(t, true, latest["B"].Equal(lastUpdate))
}

