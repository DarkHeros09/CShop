package util

import (
	"time"
)

func ParseTimeOrNil(layout string, timeStr string) *time.Time {
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		// Return nil if parsing fails
		return nil
	}
	// Return a pointer to the parsed time
	return &t
}
