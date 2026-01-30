package common_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"statements/common"
)

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
			got := common.CleanString(tt.given)
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
			amount, err := common.ParseAmount(tt.given)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				assert.InDelta(t, tt.want, amount, common.Epsilon(tt.want))
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
		{"easy", "17 Jun 2025", common.MakeDate(2025, 6, 17), ""},
		{"deceptive", "17 July 2025", common.MakeDate(2025, 7, 17), `invalid date: "17 July 2025"`},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := common.ParseDate("2 Jan 2006", tt.given)

			if tt.wantErr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			} else {
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}
