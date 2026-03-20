package common

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	amountRegexp  = regexp.MustCompile(`[^0-9.\-]+`)
	detailsRegexp = regexp.MustCompile(`\s+`)
)

// CleanString removes repeated spaces and trims ends from given string
func CleanString(s string) string {
	clean := detailsRegexp.ReplaceAllString(s, " ")
	return strings.TrimSpace(clean)
}

// ParseAmount reads a floating point value from the format "$-166.99"
func ParseAmount(amount string) (float64, error) {
	cleaned := amountRegexp.ReplaceAllString(amount, "")
	f, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount: %q", amount)
	}
	return f, nil
}

// ParseDate creates a timestamp from a date in the given format
func ParseDate(dateFormat string, date string) (time.Time, error) {
	t, err := time.Parse(dateFormat, date)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date: %q", date)
	}
	return t, nil
}

// Epsilon calculates the allowable difference between values
func Epsilon(want float64) float64 {
	return math.Nextafter(want, want+1.0) - want
}

// MakeDate constructs a new time without concern for... times
func MakeDate(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
