package pwned

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseRetryAfter(t *testing.T) {
	tests := map[string]struct {
		value string
		want  time.Duration
	}{
		"empty":        {"", 0},
		"seconds":      {"5", 5 * time.Second},
		"zero seconds": {"0", 0},
		"negative":     {"-3", 0},
		"garbage":      {"soon", 0},
		"past date":    {"Mon, 01 Jan 2000 00:00:00 GMT", 0},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, parseRetryAfter(tt.value))
		})
	}

	// A future date can only be checked as a bounded, positive delay
	t.Run("future date", func(t *testing.T) {
		future := time.Now().Add(time.Hour).UTC().Format(http.TimeFormat)
		got := parseRetryAfter(future)
		assert.Positive(t, got)
		assert.LessOrEqual(t, got, time.Hour)
	})
}

func TestRetryDelay(t *testing.T) {
	boom := errors.New("boom")

	t.Run("doubles each attempt", func(t *testing.T) {
		assert.Equal(t, 1*time.Second, retryDelay(0, boom))
		assert.Equal(t, 2*time.Second, retryDelay(1, boom))
		assert.Equal(t, 4*time.Second, retryDelay(2, boom))
		assert.Equal(t, 8*time.Second, retryDelay(3, boom))
	})

	t.Run("capped at max", func(t *testing.T) {
		assert.Equal(t, maxRetryDelay, retryDelay(4, boom))   // 16s clamps to 10s
		assert.Equal(t, maxRetryDelay, retryDelay(100, boom)) // overflow clamps too
	})

	t.Run("honors retry-after", func(t *testing.T) {
		err := &fetchError{StatusCode: http.StatusTooManyRequests, RetryAfter: 3 * time.Second}
		assert.Equal(t, 3*time.Second, retryDelay(0, err))
	})

	t.Run("retry-after capped at max", func(t *testing.T) {
		err := &fetchError{StatusCode: http.StatusTooManyRequests, RetryAfter: time.Minute}
		assert.Equal(t, maxRetryDelay, retryDelay(5, err))
	})
}
