package concatenation

// Performance experiments with various approaches to text concatenation in Go
// AMD Ryzen 9 5950X 16-Core Processor, go1.22.2 linux/amd64, Oct 2024

import (
	"strings"
)

// Naive use of + operator (264 ns/op)
// Slowest, but still suprisingly fast.
func ConcatOperator(args ...string) string {
	var s, sep string
	for _, arg := range args {
		s += sep + arg
		sep = " "
	}
	return s
}

// Explicit use of `strings.Builder` (172 ns/op)
func ConcatBuilder(args ...string) string {
	var sb strings.Builder
	var sep string
	for _, arg := range args {
		sb.WriteString(sep)
		sb.WriteString(arg)
		sep = " "
	}
	return sb.String()
}

// Use `strings.Join()` a la Python (106 ns/op)
func ConcatJoin(args ...string) string {
	return strings.Join(args, " ")
}
