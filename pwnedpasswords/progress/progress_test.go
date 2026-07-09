package progress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHumanProgress(t *testing.T) {
	full := Tally{
		FilterQueries: 130072576,
		HashQueries:   2239762,
		Found:         2239735,
		Changed:       1188872,
		Sample:        "abcdef",
	}

	t.Run("with a filter labels the leading count filtered", func(t *testing.T) {
		got := humanProgress("progress", full, true)
		assert.Equal(t, "130,072,576 filtered > 2,239,762 database reads > 1,188,872 writes > current: abcdef", got)
	})

	t.Run("without a filter shows candidates processed", func(t *testing.T) {
		got := humanProgress("progress", full, false)
		assert.Equal(t, "2,239,762 processed > 2,239,762 database reads > 1,188,872 writes > current: abcdef", got)
	})

	t.Run("omits the sample before the first database hit", func(t *testing.T) {
		c := full
		c.Sample = ""
		got := humanProgress("progress", c, true)
		assert.Equal(t, "130,072,576 filtered > 2,239,762 database reads > 1,188,872 writes", got)
	})

	t.Run("summary is prefixed with Finished", func(t *testing.T) {
		got := humanProgress("summary", full, true)
		assert.Equal(t, "Finished: 130,072,576 filtered > 2,239,762 database reads > 1,188,872 writes > current: abcdef", got)
	})
}
