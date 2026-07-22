package humanise

import (
	"math"
	"strconv"
)

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
