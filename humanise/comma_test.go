package humanise_test

import (
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
