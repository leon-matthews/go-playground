package humanise_test

import (
	"math"
	"testing"
	"time"

	"local.dev/humanise"
)

// TestDuration covers unit boundaries, the drop-to-smaller-unit quirk and specials.
func TestDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero", 0, "0 seconds"},
		{"one second", 1 * time.Second, "1 second"},
		{"two seconds", 2 * time.Second, "2 seconds"},
		{"sub-second truncates down", 1500 * time.Millisecond, "1 second"},
		{"under a second is zero", 500 * time.Millisecond, "0 seconds"},
		{"seconds", 59 * time.Second, "59 seconds"},
		{"one minute stays seconds", 60 * time.Second, "60 seconds"},
		{"two minutes", 120 * time.Second, "2 minutes"},
		{"five minutes", 300 * time.Second, "5 minutes"},
		{"one hour stays minutes", time.Hour, "60 minutes"},
		{"two hours", 2 * time.Hour, "2 hours"},
		{"one day stays hours", 24 * time.Hour, "24 hours"},
		{"two days", 48 * time.Hour, "2 days"},
		{"one week stays days", 7 * 24 * time.Hour, "7 days"},
		{"two weeks", 14 * 24 * time.Hour, "2 weeks"},
		{"docstring example", 1_000_000 * time.Second, "11 days"},
		{"one month stays weeks", 2629746 * time.Second, "4 weeks"},
		{"one year stays months", 31556952 * time.Second, "12 months"},
		{"two years", 63113904 * time.Second, "2 years"},
		{"largest span", math.MaxInt64, "292 years"},
		{"negative uses magnitude", -300 * time.Second, "5 minutes"},
		{"most negative span", math.MinInt64, "292 years"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := humanise.Duration(tt.d)
			if got != tt.want {
				t.Errorf("Duration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

// BenchmarkDuration measures formatting a span that resolves at the top unit.
func BenchmarkDuration(b *testing.B) {
	for b.Loop() {
		humanise.Duration(63113904 * time.Second)
	}
}
