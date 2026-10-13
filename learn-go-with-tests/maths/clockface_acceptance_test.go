package clockface_test

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"local.dev/clockface"
)

func TestSecondHand(t *testing.T) {
	t.Run("at midnight", func(t *testing.T) {
		now := time.Date(1337, time.January, 1, 0, 0, 0, 0, time.UTC)
		want := clockface.Point{X: 150, Y: 150 - 90}
		got := clockface.SecondHand(now)
		assert.Equal(t, want, got)
	})

	// t.Run("at 30 seconds", func(t *testing.T) {
	// 	now := time.Date(1337, time.January, 1, 0, 0, 30, 0, time.UTC)
	// 	want := clockface.Point{X: 150, Y: 150 + 90}
	// 	got := clockface.SecondHand(now)
	// 	assert.Equal(t, want, got)
	// })
}
