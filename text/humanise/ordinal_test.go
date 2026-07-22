package humanise_test

import (
	"math"
	"testing"

	"local.dev/humanise"
)

// TestOrdinal covers the last-digit suffixes and the 11-13 teens exception.
func TestOrdinal(t *testing.T) {
	tests := []struct {
		name string
		n    int64
		want string
	}{
		{"zero", 0, "0th"},
		{"first", 1, "1st"},
		{"second", 2, "2nd"},
		{"third", 3, "3rd"},
		{"fourth", 4, "4th"},
		{"eleventh", 11, "11th"},
		{"twelfth", 12, "12th"},
		{"thirteenth", 13, "13th"},
		{"fourteenth", 14, "14th"},
		{"twenty-first", 21, "21st"},
		{"twenty-second", 22, "22nd"},
		{"twenty-third", 23, "23rd"},
		{"one hundredth", 100, "100th"},
		{"hundred and first", 101, "101st"},
		{"hundred and eleventh", 111, "111th"},
		{"hundred and thirteenth", 113, "113th"},
		{"no grouping", 1000000, "1000000th"},
		{"largest", math.MaxInt64, "9223372036854775807th"},
		{"negative uses magnitude", -1, "1st"},
		{"most negative", math.MinInt64, "9223372036854775808th"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := humanise.Ordinal(tt.n)
			if got != tt.want {
				t.Errorf("Ordinal(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

// BenchmarkOrdinal measures formatting a value that hits the teens exception.
func BenchmarkOrdinal(b *testing.B) {
	for b.Loop() {
		humanise.Ordinal(112)
	}
}
