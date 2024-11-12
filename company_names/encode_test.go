package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBase36(t *testing.T) {
	trials := []struct {
		bytes  []byte
		string string
	}{
		{bytes: []byte{}, string: ""},
		{bytes: []byte{0x00}, string: "0"},
		{bytes: []byte{0x01}, string: "1"},
		{bytes: []byte{0x22}, string: "Y"},
		{bytes: []byte{0x23}, string: "Z"},
		{bytes: []byte{0x24}, string: "Z1"},
		{bytes: []byte{0x7f, 0xff, 0xff, 0xff}, string: "ZIK0ZJ"}, // Max int32
	}

	t.Run("decode", func(t *testing.T) {
		for _, trial := range trials[:1] {
			decoded := Base36Decode(trial.string)
			require.Equal(t, trial.bytes, decoded, fmt.Sprintf("Base36Encode(%q) -> %v", trial.string, decoded))
		}
	})

	t.Run("encode", func(t *testing.T) {
		for _, trial := range trials {
			encoded := Base36Encode(trial.bytes)
			require.Equal(t, trial.string, encoded, fmt.Sprintf("Base36Encode(%v) -> %q", trial.bytes, encoded))
		}
	})
}
