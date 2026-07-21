package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatSize(t *testing.T) {
	const (
		KiB = 1024
		MiB = KiB * 1024
		GiB = MiB * 1024
	)

	tests := map[string]struct {
		bytes int64
		want  string
	}{
		"zero":               {0, "0 B"},
		"one byte":           {1, "1 B"},
		"just under KiB":     {1023, "1023 B"},
		"exactly one KiB":    {KiB, "1.0 KiB"},
		"one and a half KiB": {KiB + KiB/2, "1.5 KiB"},
		"exactly one MiB":    {MiB, "1.0 MiB"},
		"one and a half MiB": {MiB + MiB/2, "1.5 MiB"},
		"exactly one GiB":    {GiB, "1.0 GiB"},
		"two and a half GiB": {GiB * 5 / 2, "2.5 GiB"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := formatSize(tc.bytes)
			assert.Equal(t, tc.want, got)
		})
	}
}
