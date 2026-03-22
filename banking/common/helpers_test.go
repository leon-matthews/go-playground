package common_test

import (
	"math"
	"time"
)

// epsilon calculates the allowable difference between values
func epsilon(want float64) float64 {
	return math.Nextafter(want, want+1.0) - want
}

// makeDate constructs a new time without concern for... times
func makeDate(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
