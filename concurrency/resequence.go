// Intended to be used after fan-out/fan-in steps in a pipeline
// See the containers folder for a priority queue example
package main

type Item[T any] struct {
	ID    int
	Value T
}

// Resequence takes an input channel of out-of-order items and returns an ordered channel.
func Resequence[T any](input <-chan Item[T]) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		pq := &PriorityQueue[T]{}
		heap.Init(pq)
		nextExpectedID := 0

		for item := range input {
			heap.Push(pq, item)

			// Release all contiguous items starting from nextExpectedID
			for pq.Len() > 0 && (*pq)[0].ID == nextExpectedID {
				nextItem := heap.Pop(pq).(Item[T])
				out <- nextItem.Value
				nextExpectedID++
			}
		}
	}()
	return out
}
