package heap_test

import (
	"cmp"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"heap/heap"
)

type item struct {
	index int
	value int
}

func (i item) compare(b item) int {
	return cmp.Compare(i.index, b.index)
}

func TestHeap(t *testing.T) {
	t.Parallel()
	unordered := []int{5, 3, 1, 2, 4}

	t.Run("new heap is empty", func(t *testing.T) {
		h := heap.NewHeap[int]()
		assert.Equal(t, 0, h.Len())
	})

	t.Run("push adds value", func(t *testing.T) {
		h := heap.NewHeap[int]()
		assert.Equal(t, 0, h.Len())
		h.Push(42)
		assert.Equal(t, 1, h.Len())
	})

	t.Run("push adds value", func(t *testing.T) {
		h := heap.NewHeap[int]()
		assert.Equal(t, 0, h.Len())
		h.Push(42)
		assert.Equal(t, 1, h.Len())
	})

	t.Run("pushing same value is okay", func(t *testing.T) {
		h := heap.NewHeap[int]()
		h.Push(42)
		h.Push(42)
		h.Push(42)
		assert.Equal(t, 3, h.Len())
		assert.Equal(t, []int{42, 42, 42}, slices.Collect(h.All()))
	})

	t.Run("peek returns smallest value", func(t *testing.T) {
		unordered := slices.Clone(unordered)
		h := heap.Heapify(unordered)
		v, ok := h.Peek()
		assert.True(t, ok)
		assert.Equal(t, 1, v)

		// Peek again
		v, ok = h.Peek()
		assert.True(t, ok)
		assert.Equal(t, 1, v)
	})

	t.Run("peek on an empty heap returns zero vale", func(t *testing.T) {
		h := heap.NewHeap[string]()
		v, ok := h.Peek()
		assert.False(t, ok)
		assert.Equal(t, "", v)
	})

	t.Run("peek does not remove value", func(t *testing.T) {
		h := heap.NewHeap[int]()
		h.Push(42)
		assert.Equal(t, 1, h.Len())

		v, ok := h.Peek()
		assert.True(t, ok)
		assert.Equal(t, 42, v)

		assert.Equal(t, 1, h.Len())
	})

	t.Run("pop returns smallest value", func(t *testing.T) {
		unordered := slices.Clone(unordered)
		h := heap.Heapify(unordered)
		v, ok := h.Pop()
		assert.True(t, ok)
		assert.Equal(t, 1, v)

		// Pop again
		v, ok = h.Pop()
		assert.True(t, ok)
		assert.Equal(t, 2, v)
	})

	t.Run("pop removes value", func(t *testing.T) {
		h := heap.NewHeap[int]()
		h.Push(42)
		assert.Equal(t, 1, h.Len())

		v, ok := h.Pop()
		assert.True(t, ok)
		assert.Equal(t, 42, v)
		assert.Equal(t, 0, h.Len())
	})

	t.Run("pop on an empty heap returns zero vale", func(t *testing.T) {
		h := heap.NewHeap[string]()
		v, ok := h.Pop()
		assert.False(t, ok)
		assert.Equal(t, "", v)
	})

	t.Run("all pops off every value, in order", func(t *testing.T) {
		unordered := slices.Clone(unordered)
		h := heap.Heapify(unordered)
		out := make([]int, 0, len(unordered))
		assert.Equal(t, 5, h.Len()) // Heap now has 5 values on it

		for v := range h.All() {
			out = append(out, v)
		}
		assert.Equal(t, []int{1, 2, 3, 4, 5}, out)
		assert.Equal(t, 0, h.Len()) // Heap is now empty
	})

	t.Run("all partially consumes heap if interrupted", func(t *testing.T) {
		unordered := slices.Clone(unordered)
		h := heap.Heapify(unordered)
		out := make([]int, 0, len(unordered))
		assert.Equal(t, "[1 2 5 3 4]", h.String())

		for v := range h.All() {
			out = append(out, v)
			if v == 3 {
				break
			}
		}
		assert.Equal(t, []int{1, 2, 3}, out)
		assert.Equal(t, 2, h.Len())
		assert.Equal(t, "[4 5]", h.String())
	})
}

func TestHeapify(t *testing.T) {
	// Build using Heapify()
	unordered := []int{5, 3, 1, 2, 4}
	heapified := heap.Heapify(slices.Clone(unordered))
	assert.Equal(t, 5, heapified.Len())

	// Build by iterative pushing
	pushed := heap.NewHeap[int]()
	for _, v := range slices.Clone(unordered) {
		pushed.Push(v)
	}
	assert.Equal(t, 5, heapified.Len())

	// The resultant heaps, built using different methods, should be equal
	assert.Equal(t, []int{1, 2, 3, 4, 5}, slices.Collect(heapified.All()))
	assert.Equal(t, []int{1, 2, 3, 4, 5}, slices.Collect(pushed.All()))
}

func TestHeapCustom(t *testing.T) {
	t.Parallel()

	t.Run("new heap is empty", func(t *testing.T) {
		h := heap.NewHeapCustom[item](item.compare)
		assert.Equal(t, 0, h.Len())
	})

	t.Run("push adds value", func(t *testing.T) {
		h := heap.NewHeapCustom[item](item.compare)
		assert.Equal(t, 0, h.Len())
		h.Push(item{1, 2})
		assert.Equal(t, 1, h.Len())
	})

	t.Run("peek does not remove value", func(t *testing.T) {
		h := heap.NewHeapCustom[item](item.compare)
		h.Push(item{2, 4})
		assert.Equal(t, 1, h.Len())

		v, ok := h.Peek()

		assert.True(t, ok)
		assert.Equal(t, item{2, 4}, v)
		assert.Equal(t, 1, h.Len())
	})

	t.Run("pop removes value", func(t *testing.T) {
		h := heap.NewHeapCustom[item](item.compare)
		h.Push(item{3, 8})
		assert.Equal(t, 1, h.Len())

		v, ok := h.Pop()

		assert.True(t, ok)
		assert.Equal(t, item{3, 8}, v)
		assert.Equal(t, 0, h.Len())
	})
}

func TestHeapifyCustom(t *testing.T) {
	unordered := makeItems(5)
	slices.Reverse(unordered)

	h := heap.HeapifyCustom(unordered, item.compare)
	assert.Equal(t, 5, h.Len())
	// Peek under the covers using the string method.
	// The first element must be the smallest.
	assert.Equal(t, "[{1 10} {2 20} {3 30} {5 50} {4 40}]", h.String())
}

func TestMakeItems(t *testing.T) {
	items := makeItems(6)
	want := []item{{1, 10}, {2, 20}, {3, 30}, {4, 40}, {5, 50}, {6, 60}}
	assert.Equal(t, want, items)
}

func TestMakeIntegers(t *testing.T) {
	items := makeIntegers(10)
	want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	assert.Equal(t, want, items)
}

// makeItems builds a slice of item where the value is index*10
// ie. {1, 2}, {2, 4}, {3, 8}...
func makeItems(count int) []item {
	unordered := make([]item, 0, count)
	for i := range count {
		index := i + 1
		unordered = append(unordered, item{index, index * 10})
	}
	return unordered
}

// makeIntegers builds a slice of integers from 1 to count
func makeIntegers(count int) []int {
	numbers := make([]int, 0, count)
	for i := range count {
		numbers = append(numbers, i+1)
	}
	return numbers
}
