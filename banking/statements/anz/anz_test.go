package anz

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// epsilon calculates the allowable difference between values
func epsilon(want float64) float64 {
	return math.Nextafter(want, want+1.0) - want
}

// makeDate constructs a new time without concern for... times
func makeDate(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func TestCleanString(t *testing.T) {
	var testcases = []struct {
		name  string
		given string
		want  string
	}{
		{"easy", "already good", "already good"},
		{"trim", " too much whitespace\n", "too much whitespace"},
		{"collapse", " \v\vfar too\t\t  \n much whitespace  \n", "far too much whitespace"},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanString(tt.given)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseAmount(t *testing.T) {
	var testcases = []struct {
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
		{"empty", "", 0, "invalid amount: \"\""},
		{"banana", "banana", 0, "invalid amount: \"banana\""},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := parseAmount(tt.given)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				assert.InDelta(t, tt.want, amount, epsilon(tt.want))
			} else {
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	var testcases = []struct {
		name    string
		given   string
		want    time.Time
		wantErr string
	}{
		{"easy", "17 Jun 2025", makeDate(2025, 6, 17), ""},
		{"deceptive", "17 July 2025", makeDate(2025, 7, 17), `invalid date: "17 July 2025"`},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.given)

			if tt.wantErr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestParseRow(t *testing.T) {
	expected := &Transaction{
		Date:      time.Date(2025, time.October, 21, 0, 0, 0, 0, time.UTC),
		Processed: time.Date(2025, time.October, 21, 0, 0, 0, 0, time.UTC),
		Account:   "4055-xxxx-1234",
		Details:   "Bob's Burgers",
		Amount:    -75.8,
	}

	t.Run("valid", func(t *testing.T) {
		row := []string{"21 Oct 2025", "21 Oct 2025", "4055-xxxx-1234", "Bob's Burgers", "$-75.80"}
		transaction, err := parseRow(row)
		assert.NoError(t, err)
		assert.Equal(t, expected, transaction)
	})

	t.Run("errors", func(t *testing.T) {
		row := []string{"", "", "", "", ""}
		wantErr := `parse row: invalid date: ""
invalid date: ""
invalid amount: ""`
		transaction, err := parseRow(row)
		assert.Nil(t, transaction)
		assert.ErrorContains(t, err, wantErr)
	})
}
