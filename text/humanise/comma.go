package humanise

import (
	"fmt"
	"strconv"
)

// Comma formats an integer with thousands separators, eg. 1234567 becomes "1,234,567".
func Comma(n int64) string {
	if n < 0 {
		// Negate in uint64 space, as the magnitude of math.MinInt64 overflows int64
		return "-" + commaUint64(uint64(-n))
	}
	return commaUint64(uint64(n))
}

// commaUint64 does the digit grouping for Comma.
func commaUint64(n uint64) string {
	if n < 1000 {
		return strconv.FormatUint(n, 10)
	}
	return commaUint64(n/1000) + "," + fmt.Sprintf("%03d", n%1000)
}
