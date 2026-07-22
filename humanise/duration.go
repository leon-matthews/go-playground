package humanise

import (
	"fmt"
	"time"
)

// secondsPerYear is the mean Gregorian year.
const secondsPerYear = 31_556_952 // 365.2425 * 24 * 60 * 60

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
// Sub-second precision is discarded and negative spans are described by their magnitude.
func Duration(d time.Duration) string {
	// Take the magnitude in seconds; dividing first keeps MinInt64 from overflowing on negate.
	seconds := int64(d / time.Second)
	if seconds < 0 {
		seconds = -seconds
	}

	// Return the first unit we have two or more of, counting from years down.
	// The count tops out near 292 years, so plain formatting never needs grouping.
	for _, unit := range durationUnits {
		if count := seconds / unit.length; count > 1 {
			return fmt.Sprintf("%d %ss", count, unit.name)
		}
	}

	// Only spans of zero or one second reach here.
	if seconds == 1 {
		return "1 second"
	}
	return fmt.Sprintf("%d seconds", seconds)
}

// Relative renders a signed time offset as an approximate human-readable phrase.
//
// A negative offset lies in the past, eg. "5 minutes ago", and a positive offset in
// the future, eg. "in 3 days"; an offset of under a second either way is "now". The
// magnitude is phrased by Duration; obtain the offset with eg. time.Until(t).
func Relative(d time.Duration) string {
	switch {
	case -time.Second < d && d < time.Second:
		return "now"
	case d < 0:
		return Duration(d) + " ago"
	default:
		return "in " + Duration(d)
	}
}
