package humanise

import "strconv"

// scale reduces value into its largest fitting unit for base, returning the
// mantissa rounded to three significant figures and the unit index (0 = first suffix).
//
// It divides value by base until the result falls below base or the last unit
// is reached, then rounds. If rounding lifts the mantissa back up to base it
// promotes to the next unit, so a value never renders as "1000" of a unit.
func scale(value, base float64, maxIndex int) (float64, int) {
	value /= base
	index := 0
	for value >= base && index < maxIndex {
		value /= base
		index++
	}
	rounded, _ := Significant(value, 3) // digits is fixed, so the error is impossible
	if rounded >= base && index < maxIndex {
		rounded /= base // rounding tipped the mantissa up a unit, eg. 999.5 -> "1MB"
		index++
	}
	return rounded, index
}

// formatMantissa renders a scaled value, trimming any trailing zeros.
func formatMantissa(m float64) string {
	return strconv.FormatFloat(m, 'g', -1, 64)
}

// magnitude returns the absolute value of n as a float64.
func magnitude(n int64) float64 {
	// Work in uint64 space, as the magnitude of math.MinInt64 overflows int64.
	u := uint64(n)
	if n < 0 {
		u = -u
	}
	return float64(u)
}
