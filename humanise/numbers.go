package humanise

import (
	"fmt"
	"math"
	"strconv"
)

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

// longScaleWords name each thousandfold multiple for Words, smallest first.
var longScaleWords = []string{"thousand", "million", "billion", "trillion", "quadrillion", "quintillion"}

// compactSuffixes abbreviate each thousandfold multiple for WordsCompact, smallest first.
var compactSuffixes = []string{"K", "M", "B", "T"}

// Words renders an integer with a long scale word, eg. 1200000 becomes "1.2 million".
//
// Values below two thousand are grouped with commas rather than worded, so an
// awkward "1.2 thousand" never appears; the thousands unit always renders as a
// whole number ("16 thousand"). Millions and above carry three significant figures.
func Words(n int64) string {
	if -2000 < n && n < 2000 {
		return Comma(n)
	}

	value := magnitude(n)
	mantissa, index := scale(value, 1000, len(longScaleWords)-1)

	var body string
	if index == 0 {
		// The thousands unit is whole-only; round from the original to avoid double rounding.
		thousands := int64(math.RoundToEven(value / 1000))
		body = strconv.FormatInt(thousands, 10) + " thousand"
	} else {
		body = formatMantissa(mantissa) + " " + longScaleWords[index]
	}

	if n < 0 {
		return "-" + body
	}
	return body
}

// WordsCompact renders an integer with a short scale suffix, eg. 1200000 becomes "1.2M".
//
// Suffixes run K, M, B, T; a value beyond a thousand trillion keeps the T suffix
// and groups its mantissa with commas ("1,000T"). Values below a thousand are
// returned as plain digits.
func WordsCompact(n int64) string {
	if -1000 < n && n < 1000 {
		return Comma(n)
	}

	mantissa, index := scale(magnitude(n), 1000, len(compactSuffixes)-1)

	var body string
	if mantissa >= 1000 {
		// A mantissa this large only occurs capped at T; group it for readability.
		body = Comma(int64(mantissa)) + compactSuffixes[index]
	} else {
		body = formatMantissa(mantissa) + compactSuffixes[index]
	}

	if n < 0 {
		return "-" + body
	}
	return body
}

// Significant rounds number to the given number of significant digits.
//
// Rounding uses round-half-to-even; NaN, ±Inf and zero pass through unchanged.
// It panics when digits is less than one. Significant is the shared rounding
// primitive the scaled formatters build on.
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
