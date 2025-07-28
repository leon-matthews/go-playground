package main

import "fmt"

func main() {
	genericLinkedList()
}

func genericLinkedList() {
	l := List[int]{}
	l.Push(21)
	l.Push(42)
	l.Push(63)
	l.Push(84)
	fmt.Println(l.AllElements())
}

// Generic linked-list
type List[T any] struct {
	head, tail *element[T]
}

type element[T any] struct {
	next *element[T]
	val  T
}

// AllElements builds and returns a new slice of list elements
func (l *List[T]) AllElements() []T {
	var elems []T
	for e := l.head; e != nil; e = e.next {
		elems = append(elems, e.val)
	}
	return elems
}

// Push adds a new element to the end of the list
func (l *List[T]) Push(v T) {
	if l.tail == nil {
		l.head = &element[T]{val: v}
		l.tail = l.head
	} else {
		l.tail.next = &element[T]{val: v}
		l.tail = l.tail.next
	}
}
