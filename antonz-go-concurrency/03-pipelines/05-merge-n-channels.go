// Merging N channels.
package main

import (
	"fmt"
	"sync"
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

// mergeN uses a single goroutines to merge N channels
func mergeN(channels ...<-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for numClosed := 0; numClosed < len(channels); {
			for i := range channels {
				select {
				case v, ok := <-channels[i]:
					if ok {
						out <- v
					} else {
						channels[i] = nil
						numClosed++
					}
				default:
				}
			}
		}
		close(out)
	}()
	return out
}

// mergeN2 uses one goroutine for each of N channels
func mergeN2(channels ...<-chan int) <-chan int {
	out := make(chan int)

	// Start a goroutine for each input channel
	var wg sync.WaitGroup
	for _, ch := range channels {
		wg.Go(func() {
			for n := range ch {
				out <- n
			}
		})
	}

	// Close output
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// solution end

func main() {
	in1 := rangeGen(11, 15)
	in2 := rangeGen(21, 25)
	in3 := rangeGen(31, 35)

	start := time.Now()
	merged := mergeN2(in1, in2, in3)
	for val := range merged {
		fmt.Print(val, " ")
	}
	fmt.Println()
	fmt.Println("Took", time.Since(start))
}
