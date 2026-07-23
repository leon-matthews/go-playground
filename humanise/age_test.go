package humanise_test

import (
	"testing"
	"time"

	"local.dev/humanise"
)

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
