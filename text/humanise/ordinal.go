package humanise

import (
	"fmt"
	"strconv"
)

// Ordinal renders a non-negative integer as its English ordinal, eg. 21 becomes "21st".
func Ordinal(n int64) (string, error) {
	if n < 0 {
		return "", fmt.Errorf("ordinal must not be negative, got %d", n)
	}

	// The 11-13 teens always take "th"; otherwise the last digit decides.
	suffix := "th"
	if lastTwo := n % 100; lastTwo < 11 || lastTwo > 13 {
		switch n % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}

	// Build the digits and suffix in one stack buffer so only the result allocates.
	var buf [21]byte // 19 digits for a positive int64, plus a two-byte suffix
	b := strconv.AppendInt(buf[:0], n, 10)
	b = append(b, suffix...)
	return string(b), nil
}
