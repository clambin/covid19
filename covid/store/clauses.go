package store

import (
	"fmt"
	"strings"
	"time"
)

func makeTimestampClause(from, to time.Time) (clause string) {
	var conditions []string
	if !from.IsZero() {
		conditions = append(conditions, fmt.Sprintf("time >= '%s'", from.Format(time.RFC3339)))
	}
	if !to.IsZero() {
		conditions = append(conditions, fmt.Sprintf("time <= '%s'", to.Format(time.RFC3339)))
	}
	if len(conditions) > 0 {
		clause = "WHERE " + strings.Join(conditions, " AND ")
	}
	return
}
