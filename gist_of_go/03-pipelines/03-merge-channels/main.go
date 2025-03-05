package main

import (
	"fmt"
	"time"
)

func main() {
	// Start two goroutines writing to two separate channels
	in1 := rangeGen(11, 15)
	in2 := rangeGen(21, 25)

	start := time.Now()
	merged := merge(in1, in2)
	for val := range merged {
		fmt.Print(val, " ")
	}
	fmt.Println()
	fmt.Println("Took", time.Since(start))
}

// merge writes values in the two input channels to the output
func merge(in1, in2 <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		select {
		case out <- <-in1:
		case out <- <-in2:
		}
	}()
	return out
}

// rangeGen SLOWLY sends values to output channel
func rangeGen(start, stop int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := start; i < stop; i++ {
			time.Sleep(50 * time.Millisecond)
			out <- i
		}
	}()
	return out
}
