package main_test

import (
	"slices"
	"testing"

	snakes "snakes-and-ladders"

	"github.com/stretchr/testify/assert"
)

func TestD6(t *testing.T) {
	const count = 1_000
	rolls := make([]int, 0, count)
	for range count {
		rolls = append(rolls, snakes.D6())
	}
	assert.Equal(t, count, len(rolls))
	assert.Equal(t, 1, slices.Min(rolls))
	assert.Equal(t, 6, slices.Max(rolls))
}

const oneMillion = 1_000_000

func BenchmarkD6(b *testing.B) {
	for b.Loop() {
		for range oneMillion {
			snakes.D6()
		}
	}
}
