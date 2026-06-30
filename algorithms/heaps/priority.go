package heaps

import (
	"cmp"
	"iter"
)

// item pairs a generic value T with its integer priority.
type item[T any] struct {
	priority int
	sequence int // insertion order, breaks ties between equal priorities
	value    T
}

func itemCompare[T any](a, b item[T]) int {
	return cmp.Or(
		cmp.Compare(a.priority, b.priority),
		cmp.Compare(a.sequence, b.sequence),
	)
}

// A PriorityQueue pops values out in ascending priority, no matter the insertion order.
//
// Values that share a priority pop in insertion order (FIFO). Built on top of a Heap,
// with the same performance guarantees.
type PriorityQueue[T any] struct {
	heap     *Heap[item[T]]
	sequence int // next sequence number, stamped into each pushed item
}

// NewPriorityQueue builds an empty priority queue of any value type.
func NewPriorityQueue[T any]() *PriorityQueue[T] {
	h := NewHeapFunc[item[T]](itemCompare)
	return &PriorityQueue[T]{heap: h}
}

// Values returns an iterator that pops values in order of ascending priority.
// The heap is consumed in the process.
func (q *PriorityQueue[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			_, v, ok := q.Pop()
			if !ok {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}

// Len returns the count of values in the queue.
func (q *PriorityQueue[T]) Len() int {
	return q.heap.Len()
}

// Push adds value v to the queue with the given priority.
func (q *PriorityQueue[T]) Push(priority int, v T) {
	q.heap.Push(item[T]{priority, q.sequence, v})
	q.sequence++
}

// Peek returns a copy of the value with the lowest priority but leaves it in place.
func (q *PriorityQueue[T]) Peek() (priority int, v T, ok bool) {
	i, ok := q.heap.Peek()
	priority, v = i.priority, i.value
	return priority, v, ok
}

// Pop returns the value with the smallest priority and removes it.
func (q *PriorityQueue[T]) Pop() (priority int, v T, ok bool) {
	i, ok := q.heap.Pop()
	priority, v = i.priority, i.value
	return priority, v, ok
}
