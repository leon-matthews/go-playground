package humanise_test

import (
	"math"
	"testing"

	"local.dev/humanise"
)

// TestOrdinal covers the last-digit suffixes and the 11-13 teens exception.
func TestOrdinal(t *testing.T) {
	tests := []struct {
		name    string
		n       int64
		want    string
		wantErr bool
	}{
		{"zero", 0, "0th", false},
		{"first", 1, "1st", false},
		{"second", 2, "2nd", false},
		{"third", 3, "3rd", false},
		{"fourth", 4, "4th", false},
		{"eleventh", 11, "11th", false},
		{"twelfth", 12, "12th", false},
		{"thirteenth", 13, "13th", false},
		{"fourteenth", 14, "14th", false},
		{"twenty-first", 21, "21st", false},
		{"twenty-second", 22, "22nd", false},
		{"twenty-third", 23, "23rd", false},
		{"one hundredth", 100, "100th", false},
		{"hundred and first", 101, "101st", false},
		{"hundred and eleventh", 111, "111th", false},
		{"hundred and thirteenth", 113, "113th", false},
		{"no grouping", 1000000, "1000000th", false},
		{"largest", math.MaxInt64, "9223372036854775807th", false},
		{"negative", -1, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := humanise.Ordinal(tt.n)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Ordinal(%d) error = %v, wantErr %v", tt.n, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
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
