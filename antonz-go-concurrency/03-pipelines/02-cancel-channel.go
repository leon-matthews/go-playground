package main

import (
	"fmt"
)

// Most common form of memory leak in Go programs is a goroutine that doesn't
// ever exit.
func main() {
	// Closing cancel signals rangeGen to exit
	cancel := make(chan struct{})
	defer close(cancel)

	numbers := rangeGen(cancel, 10, 20)
	for n := range numbers {
		fmt.Println(n)
		if n == 15 {
			break
		}
	}
}

// rangeGen sends integers from start to end (inclusive) into returned channel
func rangeGen(cancel <-chan struct{}, start int, stop int) <-chan int {
	out := make(chan int)
	current := start
	go func() {
		for {
			out <- current
			current++
			if current > stop {
				break
			}
			if <-cancel == struct{}{} {
				return
			}
		}
		close(out)
	}()
	return out
}
