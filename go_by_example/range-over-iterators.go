// Support for iterators added to Go in version 1.23, in August 2024
package main

import (
	"fmt"
	"iter"
	"math"
	"slices"
)

func main() {
	// Range over List.All()
	list := List[int]{}
	list.Push(21)
	list.Push(42)
	list.Push(63)
	list.Push(84)
	for e := range list.All() {
		fmt.Println(e)
	}

	// Build slice from iterator
	s := slices.Collect(list.All())
	fmt.Println(s)

	// Fibonacci iterator
	for n := range iFibonacci() {
		if n > math.MaxUint32 {
			break
		}

		fmt.Println(n)
	}
}

// Fibonacci iterator
func iFibonacci() iter.Seq[uint64] {
	return func(yield func(uint64) bool) {
		var a, b uint64

		a, b = 1, 1
		for {
			if !yield(a) {
				return
			}
			a, b = b, a+b
		}
	}
}

// Linked-list from generics example
type List[T any] struct {
	head, tail *element[T]
}

// Each element points to the next element, if any
type element[T any] struct {
	next *element[T]
	val  T
}

// Iterator over elements
func (l *List[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for e := l.head; e != nil; e = e.next {
			if !yield(e.val) {
				return
			}
		}
	}
}

// Push value on to end of list
func (l *List[T]) Push(v T) {
	if l.tail == nil {
		l.head = &element[T]{val: v}
		l.tail = l.head
	} else {
		l.tail.next = &element[T]{val: v}
		l.tail = l.tail.next
	}
}
