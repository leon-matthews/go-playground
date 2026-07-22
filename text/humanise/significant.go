// Package humanise formats numbers for human consumption.
package humanise

import (
	"fmt"
	"strconv"
)

// Significant rounds number to the given number of significant digits.
//
// Rounding uses round-half-to-even; NaN, ±Inf and zero pass through unchanged.
// It panics when digits is less than one.
func Significant(number float64, digits int) float64 {
	if digits < 1 {
		panic(fmt.Sprintf("humanise: digits must be at least 1, got %d", digits))
	}
	if number == 0 {
		return 0
	}
	// Round in decimal to sidestep binary scaling error and log10 edge cases.
	formatted := strconv.FormatFloat(number, 'e', digits-1, 64)
	rounded, _ := strconv.ParseFloat(formatted, 64) // FormatFloat's output always parses
	return rounded
}
