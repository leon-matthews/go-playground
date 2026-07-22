package humanise_test

import (
	"math"
	"testing"

	"local.dev/humanise"
)

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
