package heap

import (
	"cmp"
	"fmt"
	"slices"
)

// Heap is a binary heap where the smallest value is always available
type Heap[T any] struct {
	cmp    func(T, T) int
	values []T
}

// New builds an empty heap of any ordered type
func New[T cmp.Ordered]() *Heap[T] {
	return NewComparable[T](cmp.Compare[T])
}

// NewComparable builds an empty heap of any type, but requires a comparison function
func NewComparable[T any](cmp func(T, T) int) *Heap[T] {
	h := Heap[T]{
		cmp:    cmp,
		values: make([]T, 0),
	}
	return &h
}

// Heapify builds a heap from an existing slice in O(n) time
func Heapify[T cmp.Ordered](values []T) *Heap[T] {
	return HeapifyComparable(values, cmp.Compare[T])
}

// HeapifyComparable builds a heap from an existing slice with a custom comparator
func HeapifyComparable[T any](values []T, cmp func(T, T) int) *Heap[T] {
	h := &Heap[T]{
		cmp:    cmp,
		values: slices.Clone(values), // avoid mutating the caller's slice
	}
	// Start from the last non-leaf and sift down each node
	for i := len(h.values)/2 - 1; i >= 0; i-- {
		h.siftDown(i)
	}
	return h
}

// Len returns the count of values on the heap
func (h *Heap[T]) Len() int {
	return len(h.values)
}

// Peek returns the smallest value, without modifying the heap
// Uses the comma-okay idiom, returning a zero value and false if empty.
func (h *Heap[T]) Peek() (T, bool) {
	if len(h.values) == 0 {
		var zero T
		return zero, false
	}
	return h.values[0], true
}

// Pop returns the smallest value and removes it from the heap
// Uses the comma-okay idiom, returning a zero value and false if empty.
func (h *Heap[T]) Pop() (T, bool) {
	if len(h.values) == 0 {
		var zero T
		return zero, false
	}
	v := h.values[0]
	last := len(h.values) - 1
	h.values[0] = h.values[last]
	h.values = h.values[:last]
	h.siftDown(0)
	return v, true
}

// Push adds the given value to the heap
func (h *Heap[T]) Push(v T) {
	h.values = append(h.values, v)
	h.siftUp(len(h.values) - 1)
}

// Remove removes the element at index i from the heap.
// Panics if i is out of bounds.
func (h *Heap[T]) Remove(i int) T {
	if i < 0 || i >= len(h.values) {
		panic(fmt.Sprintf("heap: index %d out of bounds (len %d)", i, len(h.values)))
	}

	removed := h.values[i]
	last := len(h.values) - 1

	if i == last {
		h.values = h.values[:last]
		return removed
	}

	h.values[i] = h.values[last]
	h.values = h.values[:last]

	h.siftUp(i)
	h.siftDown(i)

	return removed
}

// String formats values as a string for debugging purposes
func (h *Heap[T]) String() string {
	return fmt.Sprint(h.values)
}

func (h *Heap[T]) siftDown(i int) {
	n := len(h.values)
	for {
		smallest, l, r := i, 2*i+1, 2*i+2
		if l < n && h.cmp(h.values[l], h.values[smallest]) < 0 {
			smallest = l
		}
		if r < n && h.cmp(h.values[r], h.values[smallest]) < 0 {
			smallest = r
		}
		if smallest == i {
			break
		}
		h.values[i], h.values[smallest] = h.values[smallest], h.values[i]
		i = smallest
	}
}

func (h *Heap[T]) siftUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if h.cmp(h.values[i], h.values[parent]) < 0 {
			h.values[i], h.values[parent] = h.values[parent], h.values[i]
			i = parent
		} else {
			break
		}
	}
}
