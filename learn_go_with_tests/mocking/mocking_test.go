package main

import (
	"bytes"
	"slices"
	"testing"
)

// Without mocking our test takes 3 seconds to run!
func TestCountdown(t *testing.T) {
	t.Run("output correct", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		sleeper := &Mock{}

		Countdown(buffer, sleeper)

		got := buffer.String()
		want := "3\n2\n1\nGo!"
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}

		if len(sleeper.Calls) != 3 {
			t.Errorf("not enough calls to sleeper, got %d want 3", len(sleeper.Calls))
		}
	})

	t.Run("call ordering", func(t *testing.T) {
		mock := &Mock{}

		Countdown(mock, mock)

		want := []string{
			write,
			sleep,
			write,
			sleep,
			write,
			sleep,
			write,
		}

		if !slices.Equal(mock.Calls, want) {
			t.Errorf("wanted calls %v got %v", want, mock.Calls)
		}
	})
}

// Keep track of both sleep and print operations
const write = "write"
const sleep = "sleep"

// Mock implements [Sleeper] and [io.Write] interfaces
type Mock struct {
	Calls []string
}

func (c *Mock) Sleep() {
	c.Calls = append(c.Calls, sleep)
}

func (c *Mock) Write(p []byte) (n int, err error) {
	c.Calls = append(c.Calls, write)
	return
}
