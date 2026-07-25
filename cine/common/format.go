// Package common holds small formatting helpers shared across the cine commands.
//
// It must sit at the end of the dependency chain: common is not allowed to import
// any other local (local.dev/cine/...) package, so it is always safe to import
// from anywhere without risking an import cycle.
package common

import (
	"fmt"
	"strconv"
	"strings"
)

// Commas formats n in base 10 with a comma between each group of three digits.
//
// For example, 1234567 becomes "1,234,567".
func Commas(n int64) string {
	s := strconv.FormatInt(n, 10)
	neg := n < 0
	if neg {
		s = s[1:]
	}

	lead := len(s) % 3
	if lead == 0 {
		lead = 3
	}

	var b strings.Builder
	b.Grow(len(s) + len(s)/3 + 1)
	if neg {
		b.WriteByte('-')
	}
	b.WriteString(s[:lead])
	for i := lead; i < len(s); i += 3 {
		b.WriteByte(',')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// Bytes formats a byte count using binary (IEC) units such as KiB and GiB.
//
// For example, 6871947674 becomes "6.4GiB".
func Bytes(n int64) string {
	const unit = 1024
	if n < unit {
		return strconv.FormatInt(n, 10) + "B"
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
