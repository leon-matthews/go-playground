package heap

// Item holds a generic value T with a sequence ID.
type Item[T any] struct {
	ID    int
	Value T
}

// PriorityQueue wraps a slice of Items for use with container/heap.
// The [Item] with the lowest ID will always be at index zero.
type PriorityQueue[T any] []Item[T]

func (pq *PriorityQueue[T]) Len() int {
	return len(pq)
}

func (pq *PriorityQueue[T]) Push(v Item[T]) {
	*pq = append(*pq, v)
}

// Peek returns a copy value with the lowest ID but leaves it in place.
func (pq *PriorityQueue[T]) Peek() Item[T] {
	return (*pq)[0]
}

// Peek removes and returns the value with the lowest ID.
func (pq *PriorityQueue[T]) Pop() Item[T] {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}
