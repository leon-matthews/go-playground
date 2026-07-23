package humanise_test

import (
	"math"
	"testing"

	"local.dev/humanise"
)

func TestFileSize_(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{"zero", 0, "0B"},
		{"bytes", 512, "512B"},
		{"just under a kilobyte", 999, "999B"},
		{"one kilobyte", 1000, "1kB"},
		{"three figures with decimals", 1234, "1.23kB"},
		{"one decimal", 4200, "4.2kB"},
		{"whole thousands", 4000, "4kB"},
		{"whole number", 42000, "42kB"},
		{"three significant figures", 456000, "456kB"},
		{"megabytes", 1500000, "1.5MB"},
		{"just under rollover", 999499, "999kB"},
		{"rollover promotes to next unit", 999999, "1MB"},
		{"terabytes", 1_000_000_000_000, "1TB"},
		{"negative bytes render by magnitude", -512, "-512B"},
		{"negative kilobytes", -4200, "-4.2kB"},
		{"most negative int64", math.MinInt64, "-9.22EB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := humanise.FileSize(tt.size); got != tt.want {
				t.Errorf("FileSize(%d) = %q, want %q", tt.size, got, tt.want)
			}
		})
	}
}

func TestFileSizeIEC_(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{"zero", 0, "0B"},
		{"bytes", 512, "512B"},
		{"just under a kibibyte", 1023, "1023B"},
		{"one kibibyte", 1024, "1KiB"},
		{"one decimal", 4200, "4.1KiB"},
		{"three significant figures", 50000, "48.8KiB"},
		{"hundreds", 500000, "488KiB"},
		{"mebibytes", 1048576, "1MiB"},
		{"negative bytes render by magnitude", -512, "-512B"},
		{"most negative int64", math.MinInt64, "-8EiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := humanise.FileSizeIEC(tt.size); got != tt.want {
				t.Errorf("FileSizeIEC(%d) = %q, want %q", tt.size, got, tt.want)
			}
		})
	}
}

// BenchmarkFileSize measures formatting a value that exercises the full unit path.
func BenchmarkFileSize(b *testing.B) {
	for b.Loop() {
		humanise.FileSize(1500000)
	}
}

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
