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

// TestRelative covers past, future, the near-zero window and the sign extremes.
func TestRelative(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero is now", 0, "now"},
		{"sub-second past is now", -999 * time.Millisecond, "now"},
		{"sub-second future is now", 999 * time.Millisecond, "now"},
		{"one second ago", -1 * time.Second, "1 second ago"},
		{"in one second", 1 * time.Second, "in 1 second"},
		{"five minutes ago", -300 * time.Second, "5 minutes ago"},
		{"in three days", 3 * 24 * time.Hour, "in 3 days"},
		{"in two years", 63113904 * time.Second, "in 2 years"},
		{"most negative offset", math.MinInt64, "292 years ago"},
		{"most positive offset", math.MaxInt64, "in 292 years"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := humanise.Relative(tt.d)
			if got != tt.want {
				t.Errorf("Relative(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

// BenchmarkRelative measures formatting a past span that resolves at a mid unit.
func BenchmarkRelative(b *testing.B) {
	for b.Loop() {
		humanise.Relative(-300 * time.Second)
	}
}

// date builds a UTC calendar date for the age tests.
func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// TestAge covers the birthday boundary, the leap-day quirk and future births.
func TestAge(t *testing.T) {
	tests := []struct {
		name  string
		born  time.Time
		today time.Time
		want  int
	}{
		{"docstring example", date(1976, time.February, 1), date(2022, time.July, 4), 46},
		{"birthday today counts", date(2000, time.June, 15), date(2020, time.June, 15), 20},
		{"day before birthday", date(2000, time.June, 15), date(2020, time.June, 14), 19},
		{"day after birthday", date(2000, time.June, 15), date(2020, time.June, 16), 20},
		{"months short of birthday", date(2000, time.December, 1), date(2020, time.June, 1), 19},
		{"born today is zero", date(2020, time.June, 15), date(2020, time.June, 15), 0},
		{"leap day, 28 Feb non-leap", date(2000, time.February, 29), date(2003, time.February, 28), 2},
		{"leap day, 1 Mar non-leap", date(2000, time.February, 29), date(2003, time.March, 1), 3},
		{"leap day, 29 Feb leap year", date(2000, time.February, 29), date(2004, time.February, 29), 4},
		{"future birth is negative", date(2030, time.June, 15), date(2020, time.June, 15), -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := humanise.Age(tt.born, tt.today)
			if got != tt.want {
				t.Errorf("Age(%s, %s) = %d, want %d",
					tt.born.Format("2006-01-02"), tt.today.Format("2006-01-02"), got, tt.want)
			}
		})
	}
}

// BenchmarkAge measures a computation that crosses the birthday boundary.
func BenchmarkAge(b *testing.B) {
	born := date(1976, time.February, 1)
	today := date(2022, time.July, 4)
	for b.Loop() {
		humanise.Age(born, today)
	}
}
