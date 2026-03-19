package heap_test

import (
	"math/rand/v2"
	"slices"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"heap/heap"
)

// BenchmarkBuildHeap compares methods of building a heap
func BenchmarkBuildHeap(b *testing.B) {
	const count = 10_000
	numbers := makeIntegers(count)
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})

	b.Run("using Heapify", func(b *testing.B) {
		var h *heap.Heap[int]
		for b.Loop() {
			b.StopTimer()
			numbers := slices.Clone(numbers)
			b.StartTimer()
			h = heap.Heapify(numbers)
		}
		v, ok := h.Peek()
		assert.True(b, ok)
		assert.Equal(b, v, 1)
	})

	b.Run("using NewHeap/Push", func(b *testing.B) {
		var h *heap.Heap[int]
		for b.Loop() {
			b.StopTimer()
			numbers := slices.Clone(numbers)
			b.StartTimer()

			h = heap.NewHeap[int]()
			for _, n := range numbers {
				h.Push(n)
			}
		}
		v, ok := h.Peek()
		assert.True(b, ok)
		assert.Equal(b, v, 1)
	})
}

// BenchmarkSort just to see how many times slower a DIY heap-sort is vs stdlib
func BenchmarkSort(b *testing.B) {
	const count = 10_000
	numbers := makeIntegers(count)
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})

	b.Run("using DIY heapsort", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			numbers := slices.Clone(numbers)
			sorted := make([]int, 0, len(numbers))
			b.StartTimer()

			h := heap.Heapify(numbers)
			for v := range h.All() {
				sorted = append(sorted, v)
			}
			require.False(b, slices.IsSorted(numbers))
			require.True(b, slices.IsSorted(sorted))
		}

	})

	b.Run("using slices.Sort()", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			numbers := slices.Clone(numbers)
			b.StartTimer()

			require.False(b, slices.IsSorted(numbers))
			slices.Sort(numbers)
			require.True(b, slices.IsSorted(numbers))
		}
	})

	b.Run("using sort.Slice()", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			numbers := slices.Clone(numbers)
			b.StartTimer()

			require.False(b, slices.IsSorted(numbers))
			sort.Slice(numbers, func(i, j int) bool { return numbers[i] < numbers[j] })
			require.True(b, slices.IsSorted(numbers))
		}
	})
}

func BenchmarkQueue(b *testing.B) {
	// Create values in random order
	items := makeItems(1_000)
	rand.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})

	var q *heap.PriorityQueue[int] // Hold on to a queue for later validation
	for b.Loop() {
		// Create new queue, then fill it n times
		q = heap.NewQueue[int]()
		for _, pair := range items {
			q.Push(pair.index, pair.value)
		}
	}

	s := slices.Collect(q.Values())
	require.True(b, slices.IsSorted(s))
}
