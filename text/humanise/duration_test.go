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
		name    string
		d       time.Duration
		want    string
		wantErr bool
	}{
		{"zero", 0, "0 seconds", false},
		{"one second", 1 * time.Second, "1 second", false},
		{"two seconds", 2 * time.Second, "2 seconds", false},
		{"sub-second truncates down", 1500 * time.Millisecond, "1 second", false},
		{"under a second is zero", 500 * time.Millisecond, "0 seconds", false},
		{"seconds", 59 * time.Second, "59 seconds", false},
		{"one minute stays seconds", 60 * time.Second, "60 seconds", false},
		{"two minutes", 120 * time.Second, "2 minutes", false},
		{"five minutes", 300 * time.Second, "5 minutes", false},
		{"one hour stays minutes", time.Hour, "60 minutes", false},
		{"two hours", 2 * time.Hour, "2 hours", false},
		{"one day stays hours", 24 * time.Hour, "24 hours", false},
		{"two days", 48 * time.Hour, "2 days", false},
		{"one week stays days", 7 * 24 * time.Hour, "7 days", false},
		{"two weeks", 14 * 24 * time.Hour, "2 weeks", false},
		{"docstring example", 1_000_000 * time.Second, "11 days", false},
		{"one month stays weeks", 2629746 * time.Second, "4 weeks", false},
		{"one year stays months", 31556952 * time.Second, "12 months", false},
		{"two years", 63113904 * time.Second, "2 years", false},
		{"largest span", math.MaxInt64, "292 years", false},
		{"negative", -1 * time.Second, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := humanise.Duration(tt.d)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Duration(%v) error = %v, wantErr %v", tt.d, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
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
