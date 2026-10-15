package pwned

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestETA(t *testing.T) {
	t.Run("estimates from the rate", func(t *testing.T) {
		remaining, ok := eta(90, 100, 10)
		assert.True(t, ok)
		assert.Equal(t, time.Second, remaining) // 10 left at 10/s
	})

	t.Run("done is a known zero", func(t *testing.T) {
		remaining, ok := eta(100, 100, 10)
		assert.True(t, ok)
		assert.Equal(t, time.Duration(0), remaining)
	})

	t.Run("past the target is a known zero", func(t *testing.T) {
		remaining, ok := eta(150, 100, 10)
		assert.True(t, ok)
		assert.Equal(t, time.Duration(0), remaining)
	})

	t.Run("stalled is unknown", func(t *testing.T) {
		remaining, ok := eta(50, 100, 0)
		assert.False(t, ok)
		assert.Equal(t, time.Duration(0), remaining)
	})
}

func TestRatePerSecond(t *testing.T) {
	t.Run("processed over the window", func(t *testing.T) {
		assert.Equal(t, 50.0, ratePerSecond(100, 2*time.Second))
	})

	t.Run("zero window is guarded", func(t *testing.T) {
		assert.Equal(t, 0.0, ratePerSecond(100, 0))
	})
}
