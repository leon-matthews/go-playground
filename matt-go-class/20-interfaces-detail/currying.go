// Currying a function reduces its argument count by one
// It is a functional programming technique
package main

import (
	"fmt"
)

func main() {
	fmt.Println(Add(10, 3))

	// adder is a function closed-over its argument
	adder := AddToA(10)
	fmt.Println(adder(3))
}

func Add(a, b int) int {
	return a + b
}

func AddToA(a int) func(int) int {
	return func(b int) int {
		return Add(a, b)
	}
}
