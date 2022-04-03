package utils_test

import (
	"github.com/clambin/covid19/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetUniqueTimestamps(t *testing.T) {
	intMap := map[time.Time]int{
		time.Date(2022, time.April, 29, 0, 0, 0, 0, time.UTC): 1,
		time.Date(2022, time.April, 30, 0, 0, 0, 0, time.UTC): 2,
		time.Date(2022, time.April, 31, 0, 0, 0, 0, time.UTC): 3,
	}

	unique := utils.GetUniqueTimestamps(intMap)

	assert.Equal(t, []time.Time{
		time.Date(2022, time.April, 29, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.April, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2022, time.April, 31, 0, 0, 0, 0, time.UTC),
	}, unique)
}
