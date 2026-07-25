package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommas(t *testing.T) {
	cases := []struct {
		name string
		in   int64
		want string
	}{
		{"zero", 0, "0"},
		{"two digits", 12, "12"},
		{"three digits", 123, "123"},
		{"four digits", 1234, "1,234"},
		{"seven digits", 1234567, "1,234,567"},
		{"principals count", 100588349, "100,588,349"},
		{"negative", -12345, "-12,345"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, Commas(tc.in))
		})
	}
}

func TestBytes(t *testing.T) {
	cases := []struct {
		name string
		in   int64
		want string
	}{
		{"bytes", 512, "512B"},
		{"kibibytes", 2048, "2.0KiB"},
		{"mebibytes", 5 * 1024 * 1024, "5.0MiB"},
		{"gibibytes", 6871947674, "6.4GiB"}, // ~6.4 * 1024^3
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, Bytes(tc.in))
		})
	}
}
