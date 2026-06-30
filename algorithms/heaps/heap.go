// Package heaps is a generic Heap and PriorityQueue implementation
package heaps

import (
	"cmp"
	"fmt"
	"iter"
)

// Heap is a generic heap implementation.
//
// Implements the classic [Heap] data structure, which efficiently (both in time
// and space) maintains a partial ordering of an underlying slice. The smallest
// value is always immediately accessible. Can be used to build priority
// queues, perform n-way (sorted) merging, and implement graph traversal algorithms.
//
// [Heap]: https://en.wikipedia.org/wiki/Heap_(data_structure)
type Heap[T any] struct {
	cmp    func(T, T) int
	values []T
}

// NewHeap builds an empty heap of any basic ordered type
func NewHeap[T cmp.Ordered]() *Heap[T] {
	return NewHeapFunc(cmp.Compare[T])
}

// NewHeapFunc builds an empty heap of any custom type, but requires a comparison function
func NewHeapFunc[T any](cmp func(T, T) int) *Heap[T] {
	h := Heap[T]{
		cmp:    cmp,
		values: make([]T, 0),
	}
	return &h
}

// Heapify builds a heap from an existing slice of basic comparable types.
//
// This is much faster than calling [NewHeap] then pushing values individually.
// The original slice is consumed in the process. Use [slices.Clone] if you need
// to keep the original unchanged.
func Heapify[T cmp.Ordered](values []T) *Heap[T] {
	return HeapifyFunc(values, cmp.Compare[T])
}

// HeapifyFunc builds a heap from an existing slice with a custom comparator
//
// This is much faster than calling [NewHeapFunc] then pushing values individually.
// The given slice is consumed in the process. Use [slices.Clone] if you need
// to keep the original slices unchanged.
func HeapifyFunc[T any](values []T, cmp func(T, T) int) *Heap[T] {
	h := &Heap[T]{
		cmp:    cmp,
		values: values,
	}
	// Start from the last non-leaf and sift down each node
	for i := len(h.values)/2 - 1; i >= 0; i-- {
		h.siftDown(i)
	}
	return h
}

// All returns an iterator that pops values off the heap in ascending order.
//
// The heap is consumed in the process.
func (h *Heap[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			v, ok := h.Pop()
			if !ok {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}

// Len returns the count of values on the heap
func (h *Heap[T]) Len() int {
	return len(h.values)
}

// Peek returns the smallest value, without modifying the heap.
func (h *Heap[T]) Peek() (value T, ok bool) {
	if len(h.values) == 0 {
		var zero T
		return zero, false
	}
	return h.values[0], true
}

// Pop returns the smallest value and removes it from the heap.
func (h *Heap[T]) Pop() (value T, ok bool) {
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

// String formats values as a string, for debugging purposes
func (h *Heap[T]) String() string {
	return fmt.Sprint(h.values)
}

// siftDown moves a value down the tree to its correct level
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

// siftUp moves a value up the tree until it reaches its correct level
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
