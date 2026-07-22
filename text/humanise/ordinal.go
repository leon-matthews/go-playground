package humanise

import "strconv"

// Ordinal renders an integer as its English ordinal, eg. 21 becomes "21st".
//
// Negative values are described by their magnitude, so -21 also becomes "21st".
func Ordinal(n int64) string {
	// Work in uint64 space, as the magnitude of math.MinInt64 overflows int64.
	u := uint64(n)
	if n < 0 {
		u = -u
	}

	// The 11-13 teens always take "th"; otherwise the last digit decides.
	suffix := "th"
	if lastTwo := u % 100; lastTwo < 11 || lastTwo > 13 {
		switch u % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}

	// Build the digits and suffix in one stack buffer so only the result allocates.
	var buf [22]byte // 20 digits for a uint64, plus a two-byte suffix
	b := strconv.AppendUint(buf[:0], u, 10)
	b = append(b, suffix...)
	return string(b)
}
