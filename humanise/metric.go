package humanise

import (
	"fmt"
	"math"
)

// siPrefixes are the SI prefixes from quecto (10^-30) up to quetta (10^30) in
// thousandfold steps; index metricCentre is the unprefixed centre.
var siPrefixes = []string{
	"q", "r", "y", "z", "a", "f", "p", "n", "µ", "m",
	"",
	"k", "M", "G", "T", "P", "E", "Z", "Y", "R", "Q",
}

// metricCentre is the index of the empty (10^0) prefix in siPrefixes.
const metricCentre = 10

// Metric renders a value with an SI prefix and unit, eg. 1500, "V" becomes "1.5 kV".
//
// The prefix is chosen so the mantissa reads between 1 and 1000, spanning quecto
// (10^-30) to quetta (10^30); values outside that range keep the extreme prefix.
// Small values take the fractional prefixes, so 0.005, "A" becomes "5 mA". The
// mantissa carries up to three significant figures and negatives keep their sign.
// It returns an error for non-finite values.
func Metric(value float64, unit string) (string, error) {
	switch {
	case math.IsNaN(value) || math.IsInf(value, 0):
		return "", fmt.Errorf("metric value must be finite, got %v", value)
	case value == 0:
		return withUnit("0", unit), nil
	}

	// Format the magnitude, then restore the sign; -5 V is a valid quantity.
	mantissa, index := scaleMetric(math.Abs(value))
	body := withUnit(formatMantissa(mantissa), siPrefixes[index]+unit)
	if value < 0 {
		return "-" + body, nil
	}
	return body, nil
}

// scaleMetric reduces value into an SI prefix, returning the mantissa rounded to
// three significant figures and the siPrefixes index.
//
// It divides or multiplies value by 1000 until it falls in [1, 1000) or an extreme
// prefix is reached. If rounding then lifts the mantissa to 1000 it promotes a prefix.
func scaleMetric(value float64) (float64, int) {
	const base = 1000.0
	index := metricCentre
	for value >= base && index < len(siPrefixes)-1 {
		value /= base
		index++
	}
	for value < 1 && index > 0 {
		value *= base
		index--
	}
	rounded := Significant(value, 3)
	if rounded >= base && index < len(siPrefixes)-1 {
		rounded /= base // rounding tipped the mantissa up a prefix, eg. 999.7k -> "1M"
		index++
	}
	return rounded, index
}

// withUnit joins a formatted number to its unit with a space, or returns the
// number alone when the unit is empty.
func withUnit(number, unit string) string {
	if unit == "" {
		return number
	}
	return number + " " + unit
}
