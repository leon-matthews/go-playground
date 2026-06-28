package quickselect

import (
	"math/rand/v2"
	"slices"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

// BenchmarkMedian compares selecting the median element via quickselect against
// fully sorting the slice and indexing it.
func BenchmarkMedian(b *testing.B) {
	const count = 1_000_000
	k := count / 2

	b.Run("using NthElement", func(b *testing.B) {
		benchmarkSelect(b, count, k)
	})

	// Shuffle once so both sort baselines race on the same input.
	numbers := makeIntegers(count)
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
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

// BenchmarkNthElementThird selects the third-smallest element, near the low end of the order.
func BenchmarkNthElementThirdPlace(b *testing.B) {
	benchmarkSelect(b, 1_000_000, 2)
}

// BenchmarkNthElementThirdToLast selects the third-largest element, near the high end.
func BenchmarkNthElementThirdToLast(b *testing.B) {
	benchmarkSelect(b, 1_000_000, 1_000_000-3)
}

// benchmarkSelect times NthElement picking index k from a shuffled slice of count integers.
func benchmarkSelect(b *testing.B, count, k int) {
	b.Helper()
	numbers := makeIntegers(count)
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})
	var got int
	for b.Loop() {
		// Clone outside the timer so only the selection is measured.
		b.StopTimer()
		values := slices.Clone(numbers)
		b.StartTimer()
		got = NthElement(values, k)
	}
	require.Equal(b, k+1, got)
}
