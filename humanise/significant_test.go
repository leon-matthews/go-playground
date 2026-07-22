package humanise_test

import (
	"fmt"
	"math"
	"testing"

	"local.dev/humanise"
)

func TestSignificant(t *testing.T) {
	tests := []struct {
		name   string
		number float64
		digits int
		want   float64
	}{
		{"rounds down", 1235, 2, 1200},
		{"rounds up", 1260, 2, 1300},
		{"tie to even rounds down", 1250, 2, 1200},
		{"tie to even rounds up", 1350, 2, 1400},
		{"carry changes magnitude", 9.99, 2, 10},
		{"single digit carries", 999, 1, 1000},
		{"small fraction", 0.0001234, 2, 0.00012},
		{"negative number", -1235, 2, -1200},
		{"exact power of ten", 1000, 2, 1000},
		{"more digits than value", 42, 5, 42},
		{"zero", 0, 2, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := humanise.Significant(tt.number, tt.digits)
			if got != tt.want {
				t.Errorf("Significant(%v, %d) = %v, want %v",
					tt.number, tt.digits, got, tt.want)
			}
		})
	}
}

// TestSignificantPanics covers the digits-out-of-range misuse cases.
func TestSignificantPanics(t *testing.T) {
	for _, digits := range []int{0, -1} {
		t.Run(fmt.Sprintf("digits %d", digits), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("Significant(100, %d) did not panic", digits)
				}
			}()
			humanise.Significant(100, digits)
		})
	}
}

// TestSignificantSpecialValues covers values that need identity comparison.
func TestSignificantSpecialValues(t *testing.T) {
	if got := humanise.Significant(math.NaN(), 2); !math.IsNaN(got) {
		t.Errorf("Significant(NaN, 2) = %v, want NaN", got)
	}
	if got := humanise.Significant(math.Inf(1), 2); !math.IsInf(got, 1) {
		t.Errorf("Significant(+Inf, 2) = %v, want +Inf", got)
	}
	if got := humanise.Significant(math.Inf(-1), 2); !math.IsInf(got, -1) {
		t.Errorf("Significant(-Inf, 2) = %v, want -Inf", got)
	}
	if got := humanise.Significant(math.Copysign(0, -1), 2); math.Signbit(got) {
		t.Errorf("Significant(-0, 2) = %v, want +0 (no sign bit)", got)
	}
}

// BenchmarkSignificant measures the format-and-parse rounding round-trip.
func BenchmarkSignificant(b *testing.B) {
	for b.Loop() {
		humanise.Significant(1234.567, 3)
	}
}
