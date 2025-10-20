package excel

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// epsilon calculates the allowable difference between values
func epsilon(want float64) float64 {
	return math.Nextafter(want, want+1.0) - want
}

func TestParseAmount(t *testing.T) {
	var tests = []struct {
		name    string
		given   string
		want    float64
		wantErr string
	}{
		// Valid
		{"easy", "1.20", 1.20, ""},
		{"dollars", "%61.99", 61.99, ""},
		{"negative", "-1.20", -1.20, ""},
		{"negative spaced", "- 1.20", -1.20, ""},
		{"negative spaced dollars", "- $1.20", -1.20, ""},
		{"large", "123_554_665.00", 123_554_665.00, ""},
		{"large negative", "-123_554_665.00", -123_554_665.00, ""},

		// Errors
		{"empty", "", 0, "parse amount: \"\""},
		{"banana", "banana", 0, "parse amount: \"banana\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check error
			amount, err := parseAmount(tt.given)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.wantErr)
			}

			// Check amount
			assert.InDelta(t, tt.want, amount, epsilon(tt.want))
		})
	}
}
