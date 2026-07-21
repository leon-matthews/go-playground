package humanise

import (
	"fmt"
	"time"
)

// secondsPerYear is the mean Gregorian year.
const secondsPerYear = int64(365.2425 * 24 * 60 * 60)

// durationUnits pairs each unit name with its length in seconds, largest first.
var durationUnits = []struct {
	name   string
	length int64
}{
	{"year", secondsPerYear},
	{"month", secondsPerYear / 12},
	{"week", 7 * 24 * 60 * 60},
	{"day", 24 * 60 * 60},
	{"hour", 60 * 60},
	{"minute", 60},
	{"second", 1},
}

// Duration renders a time span as an approximate human-readable phrase.
//
// The span is described with the largest unit of which it holds two or more,
// pluralised, eg. a 300-second span becomes "5 minutes". A span holding just one
// of a unit drops to the next unit down, so an exact week becomes "7 days".
// Sub-second precision is discarded and negative spans return an error.
func Duration(d time.Duration) (string, error) {
	if d < 0 {
		return "", fmt.Errorf("duration must not be negative, got %v", d)
	}

	// Return the first unit we have two or more of, counting from years down.
	// The count tops out near 292 years, so plain formatting never needs grouping.
	seconds := int64(d / time.Second)
	for _, unit := range durationUnits {
		if count := seconds / unit.length; count > 1 {
			return fmt.Sprintf("%d %ss", count, unit.name), nil
		}
	}

	// Only spans of zero or one second reach here.
	if seconds == 1 {
		return "1 second", nil
	}
	return fmt.Sprintf("%d seconds", seconds), nil
}
