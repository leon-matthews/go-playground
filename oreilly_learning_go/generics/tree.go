// Package tree implements a generic binary tree
package main

import (
	"cmp"
	"fmt"
)

func main() {
	t := NewTree(cmp.Compare[int])
	t.Add(13)
	t.Add(16)
	t.Add(48)
	t.Add(49)
	fmt.Printf("Contains 13 %v\n", t.Contains(13))
	fmt.Printf("Contains 17 %v\n", t.Contains(17))
}

// OrderableFunc to compares values stored in tree like `cmp.Compare`
type OrderableFunc[T any] func(t1, t2 T) int

// NewTree builds an empty tree and returns a pointer to it
func NewTree[T any](f OrderableFunc[T]) *Tree[T] {
	return &Tree[T]{
		f: f,
	}
}

// The Tree serves as the root element
type Tree[T any] struct {
	f    OrderableFunc[T]
	root *node[T]
}

// Add inserts value into tree, creating a new node if required
func (t *Tree[T]) Add(v T) {
	t.root = t.root.add(t.f, v)
}

// Contains returns true if value found in tree
func (t *Tree[T]) Contains(v T) bool {
	return t.root.contains(t.f, v)
}

// A node contains the value as well is pointers to its children
type node[T any] struct {
	val         T
	left, right *node[T]
}

func (n *node[T]) add(f OrderableFunc[T], v T) *node[T] {
	// Create node?
	if n == nil {
		return &node[T]{val: v}
	}

	// Add to left or right
	switch r := f(v, n.val); {
	case r <= -1:
		n.left = n.left.add(f, v)
	case r >= 1:
		n.left = n.right.add(f, v)
	}

	// Value already in this node
	return n
}

func (n *node[T]) contains(f OrderableFunc[T], v T) bool {
	if n == nil {
		return false
	}

	// Add to left or right
	switch r := f(v, n.val); {
	case r <= -1:
		return n.left.contains(f, v)
	case r >= 1:
		return n.right.contains(f, v)
	}

	// Value is in this node
	return true
}
