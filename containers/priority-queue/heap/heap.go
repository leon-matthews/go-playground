package heap

import (
	"cmp"
	"fmt"
)

// Heap is a binary heap where the smallest value is always available
type Heap[T any] struct {
	cmp    func(T, T) int
	values []T
}

// New builds a heap of any ordered type
func New[T cmp.Ordered]() *Heap[T] {
	return NewComparable[T](cmp.Compare[T])
}

// NewComparable builds a heap of any type, but requires a comparison function
func NewComparable[T any](cmp func(T, T) int) *Heap[T] {
	h := Heap[T]{
		cmp:    cmp,
		values: make([]T, 0),
	}
	return &h
}

// Len returns the count of values on the heap
func (h *Heap[T]) Len() int {
	return len(h.values)
}

// Peek returns the smallest value, without modifying the heap
func (h *Heap[T]) Peek() T {
	return h.values[0]
}

// Pop returns the smallest value and removes it from the heap
func (h *Heap[T]) Pop() T {
	v := h.values[0]
	h.values = h.values[1:]
	return v
}

// Push adds the given value to the heap
func (h *Heap[T]) Push(v T) {
	h.values = append(h.values, v)
}

// String formats values as a string for debugging purposes
func (h *Heap[T]) String() string {
	return fmt.Sprint(h.values)
}
