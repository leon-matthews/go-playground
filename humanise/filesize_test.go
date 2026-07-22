package humanise_test

import (
	"math"
	"testing"

	"local.dev/humanise"
)

func TestFileSize_(t *testing.T) {
	tests := []struct {
		name    string
		size    float64
		want    string
		wantErr bool
	}{
		{"zero", 0, "0B", false},
		{"bytes", 512, "512B", false},
		{"fractional bytes round", 512.7, "513B", false},
		{"just under a kilobyte", 999, "999B", false},
		{"rounds up to a kilobyte", 999.7, "1kB", false},
		{"one kilobyte", 1000, "1kB", false},
		{"three figures with decimals", 1234, "1.23kB", false},
		{"one decimal", 4200, "4.2kB", false},
		{"whole thousands", 4000, "4kB", false},
		{"whole number", 42000, "42kB", false},
		{"three significant figures", 456000, "456kB", false},
		{"megabytes", 1500000, "1.5MB", false},
		{"just under rollover", 999499, "999kB", false},
		{"rollover promotes to next unit", 999999, "1MB", false},
		{"terabytes", 1e12, "1TB", false},
		{"negative", -1, "", true},
		{"not a number", math.NaN(), "", true},
		{"infinite", math.Inf(1), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := humanise.FileSize(tt.size)
			if (err != nil) != tt.wantErr {
				t.Fatalf("FileSize(%v) error = %v, wantErr %v", tt.size, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("FileSize(%v) = %q, want %q", tt.size, got, tt.want)
			}
		})
	}
}

func TestFileSizeIEC_(t *testing.T) {
	tests := []struct {
		name    string
		size    float64
		want    string
		wantErr bool
	}{
		{"zero", 0, "0B", false},
		{"bytes", 512, "512B", false},
		{"just under a kibibyte", 1023, "1023B", false},
		{"rounds up to a kibibyte", 1023.7, "1KiB", false},
		{"one kibibyte", 1024, "1KiB", false},
		{"one decimal", 4200, "4.1KiB", false},
		{"three significant figures", 50000, "48.8KiB", false},
		{"hundreds", 500000, "488KiB", false},
		{"mebibytes", 1048576, "1MiB", false},
		{"negative", -1, "", true},
		{"not a number", math.NaN(), "", true},
		{"infinite", math.Inf(-1), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := humanise.FileSizeIEC(tt.size)
			if (err != nil) != tt.wantErr {
				t.Fatalf("FileSizeIEC(%v) error = %v, wantErr %v", tt.size, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("FileSizeIEC(%v) = %q, want %q", tt.size, got, tt.want)
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
