package main_test

import (
	"fmt"
)

// Most common form of memory leak in Go programs is a goroutine that doesn't exit.
func main() {
	// Channel closed, generator's coroutine exits
	numbers := RangeGen(1, 5)
	for n := range numbers {
		fmt.Println(n)
	}

	// Exiting early leaves the generator's goroutine running!
	for n := range RangeGen(1, 10) {
		fmt.Println(n)
		if n == 3 {
			break
		}
	}
}

// RangeGen sends integers from start to end (inclusive) into returned channel
func RangeGen(start int, stop int) <-chan int {
	out := make(chan int)
	current := start
	go func() {
		for {
			out <- current
			current++
			if current > stop {
				break
			}
		}
		close(out)
	}()
	return out
}
