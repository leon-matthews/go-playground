package quickselect_test

import (
	"math/rand/v2"
	"slices"
	"sort"
	"testing"

	"quickselect/quickselect"

	"github.com/stretchr/testify/require"
)

// BenchmarkNthElement compares selecting one element via quickselect against
// fully sorting the slice and indexing it, for the headline median case.
func BenchmarkNthElement(b *testing.B) {
	const count = 1_000_000
	k := count / 2
	numbers := makeIntegers(count)
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})

	b.Run("using NthElement", func(b *testing.B) {
		var got int
		for b.Loop() {
			b.StopTimer()
			values := slices.Clone(numbers)
			b.StartTimer()
			got = quickselect.NthElement(values, k)
		}
		require.Equal(b, k+1, got)
	})

	b.Run("using slices.Sort", func(b *testing.B) {
		var got int
		for b.Loop() {
			b.StopTimer()
			values := slices.Clone(numbers)
			b.StartTimer()
			slices.Sort(values)
			got = values[k]
		}
		require.Equal(b, k+1, got)
	})

	b.Run("using sort.Slice", func(b *testing.B) {
		var got int
		for b.Loop() {
			b.StopTimer()
			values := slices.Clone(numbers)
			b.StartTimer()
			sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
			got = values[k]
		}
		require.Equal(b, k+1, got)
	})
}
