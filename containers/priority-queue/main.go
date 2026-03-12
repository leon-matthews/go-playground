package main

import (
	"fmt"

	"priority/heap"
)

func main() {
	h := heap.New[int]()
	h.Push(42)
	fmt.Println(h.Len())
	fmt.Println(h)
}
