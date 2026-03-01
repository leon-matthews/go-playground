// Merging N channels.
package main

import (
	"fmt"
	"time"
)

// rangeGen sends numbers from start to stop-1 to the channel.
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

// solution start

// merge selects numbers from input channels and sends them to the output.
func merge(channels ...<-chan int) <-chan int {
	// ...
}

// solution end

func main() {
	in1 := rangeGen(11, 15)
	in2 := rangeGen(21, 25)
	in3 := rangeGen(31, 35)

	start := time.Now()
	merged := merge(in1, in2, in3)
	for val := range merged {
		fmt.Print(val, " ")
	}
	fmt.Println()
	fmt.Println("Took", time.Since(start))
}
