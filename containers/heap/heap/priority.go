package heap

import "cmp"

// item holds a generic value T with a sequence id.
type item[T any] struct {
	id    int
	value T
}

func itemCompare[T any](a, b item[T]) int {
	return cmp.Compare(a.id, b.id)
}

// A PriorityQueue pops values out in ascending order, no matter the insertion order.
// It is built on top of a Heap, with the same performance guarantees.
type PriorityQueue[T cmp.Ordered] struct {
	heap *Heap[item[T]]
}

func NewQueue[T cmp.Ordered]() *PriorityQueue[T] {
	h := NewHeapCustom[item[T]](itemCompare)
	return &PriorityQueue[T]{heap: h}
}

func (pq *PriorityQueue[T]) Len() int {
	return pq.heap.Len()
}

func (pq *PriorityQueue[T]) Push(id int, v T) {
	pq.heap.Push(item[T]{id, v})
}

// Peek returns a copy of value with the lowest ID but leaves it in place.
func (pq *PriorityQueue[T]) Peek() (id int, v T, ok bool) {
	i, ok := pq.heap.Peek()
	id, v = i.id, i.value
	return id, v, ok
}

// Pop returns the value with the smallest id and removes it
func (pq *PriorityQueue[T]) Pop() (id int, v T, ok bool) {
	i, ok := pq.heap.Pop()
	id, v = i.id, i.value
	return id, v, ok
}
