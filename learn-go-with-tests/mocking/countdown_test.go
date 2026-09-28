package main

import (
	"bytes"
	"testing"
)

// SleeperMock keeps track of how many times it was called, and quickly.
type SleeperMock struct {
    Calls int
}

func (s *SleeperMock) Sleep() {
    s.Calls += 1
}

func TestCountdown(t *testing.T) {
	buffer := &bytes.Buffer{}
	sleeper := &SleeperMock{}

	Countdown(buffer, sleeper)

    // Check output
	got := buffer.String()
	want := `3
2
1
Go!`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// Spy on sleeper
	if sleeper.Calls != 3 {
        t.Errorf("got %d calls, want %d", sleeper.Calls, 3)
	}
}
