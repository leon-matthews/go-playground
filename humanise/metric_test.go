package humanise_test

import (
	"math"
	"testing"

	"local.dev/humanise"
)

func TestMetric(t *testing.T) {
	tests := []struct {
		name    string
		value   float64
		unit    string
		want    string
		wantErr bool
	}{
		{"kilo", 1500, "V", "1.5 kV", false},
		{"mega", 2e8, "W", "200 MW", false},
		{"bare unit", 500, "V", "500 V", false},
		{"three significant figures", 1234, "Hz", "1.23 kHz", false},
		{"milli", 5e-3, "A", "5 mA", false},
		{"micro", 220e-6, "F", "220 µF", false},
		{"nano", 3e-9, "s", "3 ns", false},
		{"negative kilo", -1500, "V", "-1.5 kV", false},
		{"negative milli", -5e-3, "A", "-5 mA", false},
		{"zero with unit", 0, "V", "0 V", false},
		{"zero without unit", 0, "", "0", false},
		{"prefix without unit", 1500, "", "1.5 k", false},
		{"bare without unit", 500, "", "500", false},
		{"rounds up into the centre", 999.7, "V", "1 kV", false},
		{"rounds up a prefix", 999700, "W", "1 MW", false},
		{"clamps at quetta", 1e33, "W", "1000 QW", false},
		{"clamps at quecto", 1e-33, "s", "0.001 qs", false},
		{"not a number", math.NaN(), "V", "", true},
		{"positive infinity", math.Inf(1), "V", "", true},
		{"negative infinity", math.Inf(-1), "V", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := humanise.Metric(tt.value, tt.unit)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Metric(%v, %q) error = %v, wantErr %v", tt.value, tt.unit, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("Metric(%v, %q) = %q, want %q", tt.value, tt.unit, got, tt.want)
			}
		})
	}
}

// BenchmarkMetric measures formatting a small value that exercises the multiply path.
func BenchmarkMetric(b *testing.B) {
	for b.Loop() {
		humanise.Metric(220e-6, "F")
	}
}
