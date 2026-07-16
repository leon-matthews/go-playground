package progress

import (
	"testing"
	"time"

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
	prev := Tally{
		FilterQueries: 129072576, // 1,000,000 candidates over 10s -> 100,000/s
		HashQueries:   2039762,   //   200,000 db reads over 10s    ->  20,000/s
	}
	const since = 10 * time.Second

	t.Run("with a filter counts filter lookups as candidates", func(t *testing.T) {
		got := humanProgress("progress", full, prev, since, true)
		assert.Equal(t, "130,072,576 candidates (100,000/s) > 20,000 db reads/s > 1,188,872 writes > found: abcdef", got)
	})

	t.Run("without a filter counts hash lookups as candidates", func(t *testing.T) {
		got := humanProgress("progress", full, prev, since, false)
		assert.Equal(t, "2,239,762 candidates (20,000/s) > 20,000 db reads/s > 1,188,872 writes > found: abcdef", got)
	})

	t.Run("omits the sample before the first database hit", func(t *testing.T) {
		c := full
		c.Sample = ""
		got := humanProgress("progress", c, prev, since, true)
		assert.Equal(t, "130,072,576 candidates (100,000/s) > 20,000 db reads/s > 1,188,872 writes", got)
	})

	t.Run("shows zero rates when no time has passed", func(t *testing.T) {
		got := humanProgress("progress", full, prev, 0, true)
		assert.Equal(t, "130,072,576 candidates (0/s) > 0 db reads/s > 1,188,872 writes > found: abcdef", got)
	})

	t.Run("summary shows totals only, without rate or sample", func(t *testing.T) {
		got := humanProgress("summary", full, prev, since, true)
		assert.Equal(t, "Finished: 130,072,576 candidates > 2,239,762 db reads > 1,188,872 writes", got)
	})
}
