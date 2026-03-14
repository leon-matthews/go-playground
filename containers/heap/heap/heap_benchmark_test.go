package heap_test

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"heap/heap"
)

// BenchmarkBuildHeap compares methods of building a heap
func BenchmarkBuildHeap(b *testing.B) {
	const count = 10_000
	numbers := makeIntegers(count)
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})

	b.Run("Build using heapify", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			numbers := slices.Clone(numbers)
			b.StartTimer()

			h := heap.Heapify(numbers)
			v, ok := h.Peek()
			assert.True(b, ok)
			assert.Equal(b, v, 1)
		}
	})

	b.Run("Build using new/push", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			numbers := slices.Clone(numbers)
			b.StartTimer()

			h := heap.New[int]()
			for _, n := range numbers {
				h.Push(n)
			}
			v, ok := h.Peek()
			assert.True(b, ok)
			assert.Equal(b, v, 1)
		}
	})
}

// BenchmarkSort compares a DIY heap-sort with [slices.Sort]
func BenchmarkSort(b *testing.B) {
	const count = 10_000
	numbers := makeIntegers(count)
	rand.Shuffle(len(numbers), func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})

	b.Run("Sort using DIY heapsort", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			numbers := slices.Clone(numbers)
			sorted := make([]int, 0, len(numbers))
			b.StartTimer()

			h := heap.Heapify(numbers)
			for v := range h.All() {
				sorted = append(sorted, v)
			}
			assert.True(b, slices.IsSorted(sorted))
		}
	})

	b.Run("Sort using slices.Sort()", func(b *testing.B) {
		for b.Loop() {
			b.StopTimer()
			numbers := slices.Clone(numbers)
			b.StartTimer()

			slices.Sort(numbers)
			assert.True(b, slices.IsSorted(numbers))
		}
	})
}
