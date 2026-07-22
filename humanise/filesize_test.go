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
