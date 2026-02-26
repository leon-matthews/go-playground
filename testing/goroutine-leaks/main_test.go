package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

// You can either use a TestMain like here, or add a line to each test:
//   defer goleak.VerifyNone(t)
//
// Install in the usual way:
//   go get -u go.uber.org/goleak
//
// Caveat: For tests that use t.Parallel, goleak does not know how to
// distinguish a leaky goroutine from tests that have not finished running.
// https://github.com/uber-go/goleak
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestRangeGen(t *testing.T) {
	generator := RangeGen(-5, 5)
	numbers := make([]int, 0)
	for n := range generator {
		numbers = append(numbers, n)
	}
	want := []int{-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5}
	assert.Equal(t, want, numbers)
}

func TestRangeGenLeaky(t *testing.T) {
	generator := RangeGen(-5, 5)
	numbers := make([]int, 0)
	for n := range generator {
		if n == 0 {
			break
		}
		numbers = append(numbers, n)
	}
	want := []int{-5, -4, -3, -2, -1}
	assert.Equal(t, want, numbers)
}
