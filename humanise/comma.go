package humanise

// Comma formats an integer with thousands separators, eg. 1234567 becomes "1,234,567".
func Comma(n int64) string {
	return group(n, ',')
}

// Underscore formats an integer as a Go numeric literal, eg. 1234567 becomes "1_234_567".
func Underscore(n int64) string {
	return group(n, '_')
}

// group formats n in decimal, writing sep between each three-digit group.
func group(n int64, sep byte) string {
	// Work in uint64 space, as the magnitude of math.MinInt64 overflows int64
	u := uint64(n)
	if n < 0 {
		u = -u
	}

	// Fill the buffer backwards; 26 bytes fit 19 digits, 6 separators and a sign
	var buf [26]byte
	i := len(buf)
	for digits := 0; ; {
		i--
		buf[i] = byte('0' + u%10)
		u /= 10
		if u == 0 {
			break
		}
		digits++
		if digits%3 == 0 {
			i--
			buf[i] = sep
		}
	}
	if n < 0 {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
