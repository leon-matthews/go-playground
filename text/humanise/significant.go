// Package humanise formats numbers for human consumption.
package humanise

import (
	"fmt"
	"strconv"
)

// Significant rounds number to the given number of significant digits.
//
// Rounding uses round-half-to-even; NaN, ±Inf and zero pass through unchanged.
// It returns an error when digits is less than one.
func Significant(number float64, digits int) (float64, error) {
	if digits < 1 {
		return 0, fmt.Errorf("digits must be at least 1, got %d", digits)
	}
	if number == 0 {
		return 0, nil
	}
	// Round in decimal to sidestep binary scaling error and log10 edge cases.
	formatted := strconv.FormatFloat(number, 'e', digits-1, 64)
	return strconv.ParseFloat(formatted, 64)
}
