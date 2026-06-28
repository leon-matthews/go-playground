package heap

import (
	"cmp"
	"iter"
)

// item holds a generic value T with a sequence id.
type item[T any] struct {
	id    int
	value T
}

func itemCompare[T any](a, b item[T]) int {
	return cmp.Compare(a.id, b.id)
}

// A PriorityQueue pops values out in ascending ID, no matter the insertion order.
// It is built on top of a Heap, with the same performance guarantees.
type PriorityQueue[T any] struct {
	heap *Heap[item[T]]
}

func NewQueue[T any]() *PriorityQueue[T] {
	h := NewHeapCustom[item[T]](itemCompare)
	return &PriorityQueue[T]{heap: h}
}

// Values returns an iterator that pops values in order of ascending id
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

func (q *PriorityQueue[T]) Len() int {
	return q.heap.Len()
}

func (q *PriorityQueue[T]) Push(id int, v T) {
	q.heap.Push(item[T]{id, v})
}

// Peek returns a copy of value with the lowest ID but leaves it in place.
func (q *PriorityQueue[T]) Peek() (id int, v T, ok bool) {
	i, ok := q.heap.Peek()
	id, v = i.id, i.value
	return id, v, ok
}

// Pop returns the value with the smallest id and removes it
func (q *PriorityQueue[T]) Pop() (id int, v T, ok bool) {
	i, ok := q.heap.Pop()
	id, v = i.id, i.value
	return id, v, ok
}
