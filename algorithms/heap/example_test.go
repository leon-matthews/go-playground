package heap_test

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"local.dev/heap"
)

// Example_sorting builds a heap, peeks the smallest value, and drains it in order.
func Example_sorting() {
	h := heap.Heapify([]int{5, 3, 1, 4, 2})
	smallest, _ := h.Peek()
	fmt.Println("smallest:  ", smallest)
	fmt.Println("ascending: ", slices.Collect(h.All()))

	// A descending comparator builds a max-heap.
	descending := func(a, b int) int { return cmp.Compare(b, a) }
	maxHeap := heap.HeapifyFunc([]int{5, 3, 1, 4, 2}, descending)
	fmt.Println("descending:", slices.Collect(maxHeap.All()))
	// Output:
	// smallest:   1
	// ascending:  [1 2 3 4 5]
	// descending: [5 4 3 2 1]
}

// Example_topN keeps the n largest values of a stream using a min-heap of size n.
func Example_topN() {
	const n = 3
	stream := []int{7, 3, 11, 1, 9, 2, 12, 5, 8, 4, 10, 6}

	h := heap.NewHeap[int]()
	for _, v := range stream {
		h.Push(v)
		if h.Len() > n {
			h.Pop() // evict the smallest, keeping the n largest seen so far
		}
	}

	fmt.Println("top 3 largest:", slices.Collect(h.All()))
	// Output:
	// top 3 largest: [10 11 12]
}

// Example_priorityQueue pops values by ascending priority, whatever the push order.
func Example_priorityQueue() {
	q := heap.NewPriorityQueue[string]()
	q.Push(3, "black")
	q.Push(1, "Sphinx")
	q.Push(4, "quartz")
	q.Push(2, "of")
	q.Push(6, "my")
	q.Push(5, "judge")
	q.Push(7, "vow")

	fmt.Println(strings.Join(slices.Collect(q.Values()), " "))
	// Output:
	// Sphinx of black quartz judge my vow
}
