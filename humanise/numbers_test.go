package humanise_test

import (
	"fmt"
	"math"
	"testing"

	"local.dev/humanise"
)

// TestComma checks thousands separators across magnitudes and signs.
func TestComma(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{1234567, "1,234,567"},
		{-42, "-42"},
		{-1234567, "-1,234,567"},
		{math.MaxInt64, "9,223,372,036,854,775,807"},
		{math.MinInt64, "-9,223,372,036,854,775,808"},
	}
	for _, test := range tests {
		if got := humanise.Comma(test.n); got != test.want {
			t.Errorf("comma(%d) = %q, want %q", test.n, got, test.want)
		}
	}
}

// TestUnderscore checks Go literal grouping across magnitudes and signs.
func TestUnderscore(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1_000"},
		{1234567, "1_234_567"},
		{-42, "-42"},
		{-1234567, "-1_234_567"},
		{math.MaxInt64, "9_223_372_036_854_775_807"},
		{math.MinInt64, "-9_223_372_036_854_775_808"},
	}
	for _, test := range tests {
		if got := humanise.Underscore(test.n); got != test.want {
			t.Errorf("underscore(%d) = %q, want %q", test.n, got, test.want)
		}
	}
}

// BenchmarkComma measures formatting a worst-case 19-digit value.
func BenchmarkComma(b *testing.B) {
	for b.Loop() {
		humanise.Comma(math.MaxInt64)
	}
}

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

func TestWords_(t *testing.T) {
	tests := []struct {
		name string
		n    int64
		want string
	}{
		{"zero", 0, "0"},
		{"small", 999, "999"},
		{"under two thousand", 1200, "1,200"},
		{"just under threshold", 1999, "1,999"},
		{"two thousand", 2000, "2 thousand"},
		{"whole thousands", 3000, "3 thousand"},
		{"coarse round down", 3400, "3 thousand"},
		{"half rounds to even", 2500, "2 thousand"},
		{"fractional rounds up", 15500, "16 thousand"},
		{"hundreds of thousands", 420000, "420 thousand"},
		{"rounds to whole thousands", 123400, "123 thousand"},
		{"promotes to million", 999999, "1 million"},
		{"one million", 1000000, "1 million"},
		{"decimal million", 1200000, "1.2 million"},
		{"three figure million", 123000000, "123 million"},
		{"one billion", 1000000000, "1 billion"},
		{"promotes to billion", 999500000, "1 billion"},
		{"quintillion", 9220000000000000000, "9.22 quintillion"},
		{"negative million", -1200000, "-1.2 million"},
		{"negative thousands", -2000, "-2 thousand"},
		{"minimum int64", math.MinInt64, "-9.22 quintillion"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := humanise.Words(tt.n); got != tt.want {
				t.Errorf("Words(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

func TestWordsCompact_(t *testing.T) {
	tests := []struct {
		name string
		n    int64
		want string
	}{
		{"zero", 0, "0"},
		{"small", 999, "999"},
		{"one thousand", 1000, "1K"},
		{"decimal thousands", 1500, "1.5K"},
		{"hundreds of thousands", 420000, "420K"},
		{"million", 1200000, "1.2M"},
		{"promotes to billion", 999500000, "1B"},
		{"trillion", 1000000000000, "1T"},
		{"capped at trillion grows", 1000000000000000, "1,000T"},
		{"negative million", -1200000, "-1.2M"},
		{"maximum int64", math.MaxInt64, "9,220,000T"},
		{"minimum int64", math.MinInt64, "-9,220,000T"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := humanise.WordsCompact(tt.n); got != tt.want {
				t.Errorf("WordsCompact(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

// BenchmarkWords measures formatting a value that reaches the millions unit.
func BenchmarkWords(b *testing.B) {
	for b.Loop() {
		humanise.Words(1200000)
	}
}

// BenchmarkWordsCompact measures formatting a value that reaches the millions unit.
func BenchmarkWordsCompact(b *testing.B) {
	for b.Loop() {
		humanise.WordsCompact(1200000)
	}
}

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
