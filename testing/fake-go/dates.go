package fake

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Duration of each unit in seconds, using the Gregorian calendar year.
const (
	secondsPerHour  = 3600.0
	secondsPerDay   = secondsPerHour * 24
	secondsPerYear  = secondsPerDay * 365.2425
	secondsPerMonth = secondsPerYear / 12
	secondsPerWeek  = secondsPerYear / 52
)

// durationUnits maps a unit's first letter to its length in seconds.
//
// Note that 'm' is a month, not a minute, matching the Python original.
var durationUnits = map[byte]float64{
	'd': secondsPerDay,
	'h': secondsPerHour,
	'm': secondsPerMonth,
	'w': secondsPerWeek,
	'y': secondsPerYear,
}

// RelativeTime returns a time relative to now.
//
// The value is either "now" or one or more signed duration pairs such as
// "+3 days", "-40 years", or "2y4w7d". Units are d, h, w, m (month), and y,
// matched by first letter. It returns an error for an unparseable value.
func (f *Faker) RelativeTime(value string) (time.Time, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	now := time.Now()
	if value == "now" {
		return now, nil
	}
	delta, err := parsePairs(value)
	if err != nil {
		return time.Time{}, err
	}
	return now.Add(delta), nil
}

// Between returns a random time in the half-open range [start, end).
//
// It panics if start is not before end.
func (f *Faker) Between(start, end time.Time) time.Time {
	if !start.Before(end) {
		panic("fake: start must come before end")
	}
	offset := time.Duration(f.rng.Float64() * float64(end.Sub(start)))
	return start.Add(offset)
}

// parsePairs sums any number of signed number-and-unit duration pairs.
func parsePairs(value string) (time.Duration, error) {
	value = strings.ReplaceAll(value, " ", "")
	seconds := 0.0
	for i := 0; i < len(value); {
		start := i
		for i < len(value) && !isLower(value[i]) {
			i++
		}
		number := value[start:i]

		start = i
		for i < len(value) && isLower(value[i]) {
			i++
		}
		unit := value[start:i]

		quantity, err := strconv.ParseFloat(number, 64)
		if err != nil {
			return 0, fmt.Errorf("fake: invalid duration quantity: %q", number)
		}
		if unit == "" {
			return 0, fmt.Errorf("fake: empty duration unit")
		}
		unitSeconds, ok := durationUnits[unit[0]]
		if !ok {
			return 0, fmt.Errorf("fake: invalid duration unit: %q", unit)
		}
		seconds += quantity * unitSeconds
	}
	return time.Duration(seconds * float64(time.Second)), nil
}

// isLower reports whether b is an ASCII lower-case letter.
func isLower(b byte) bool {
	return b >= 'a' && b <= 'z'
}
