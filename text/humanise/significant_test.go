package humanise

import (
	"math"
	"testing"
)

func TestSignificant(t *testing.T) {
	tests := []struct {
		name    string
		number  float64
		digits  int
		want    float64
		wantErr bool
	}{
		{"rounds down", 1235, 2, 1200, false},
		{"rounds up", 1260, 2, 1300, false},
		{"tie to even rounds down", 1250, 2, 1200, false},
		{"tie to even rounds up", 1350, 2, 1400, false},
		{"carry changes magnitude", 9.99, 2, 10, false},
		{"single digit carries", 999, 1, 1000, false},
		{"small fraction", 0.0001234, 2, 0.00012, false},
		{"negative number", -1235, 2, -1200, false},
		{"exact power of ten", 1000, 2, 1000, false},
		{"more digits than value", 42, 5, 42, false},
		{"zero", 0, 2, 0, false},
		{"zero digits is an error", 100, 0, 0, true},
		{"negative digits is an error", 100, -1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Significant(tt.number, tt.digits)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Significant(%v, %d) error = %v, wantErr %v",
					tt.number, tt.digits, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("Significant(%v, %d) = %v, want %v",
					tt.number, tt.digits, got, tt.want)
			}
		})
	}
}

// TestSignificantSpecialValues covers values that need identity comparison.
func TestSignificantSpecialValues(t *testing.T) {
	if got, _ := Significant(math.NaN(), 2); !math.IsNaN(got) {
		t.Errorf("Significant(NaN, 2) = %v, want NaN", got)
	}
	if got, _ := Significant(math.Inf(1), 2); !math.IsInf(got, 1) {
		t.Errorf("Significant(+Inf, 2) = %v, want +Inf", got)
	}
	if got, _ := Significant(math.Inf(-1), 2); !math.IsInf(got, -1) {
		t.Errorf("Significant(-Inf, 2) = %v, want -Inf", got)
	}
	if got, _ := Significant(math.Copysign(0, -1), 2); math.Signbit(got) {
		t.Errorf("Significant(-0, 2) = %v, want +0 (no sign bit)", got)
	}
}
