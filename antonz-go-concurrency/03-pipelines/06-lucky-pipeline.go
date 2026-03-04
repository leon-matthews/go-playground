package main

import (
	"fmt"
	"sync"
)

// integers generates the numbers from start to stop, inclusive.
func integers(start int, stop int) <-chan int {
	out := make(chan int)
	go func() {
		for i := start; i <= stop; i++ {
			out <- i
		}
		close(out)
	}()
	return out
}

// findLucky passes only the luckiest numbers on to the output channel
func findLucky(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if n%7 == 0 && n%13 != 0 {
				out <- n
			}
		}
	}()
	return out
}

// merge combines N input channels into one output channel using N goroutines
func merge(channels []<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	for i := range channels {
		wg.Go(func() {
			for n := range channels[i] {
				out <- n
			}
		})
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	const concurrency = 4
	numbers := integers(1, 100)

	// Create N lucky number finders
	luckyNumbers := make([]<-chan int, concurrency)
	for i := range len(luckyNumbers) {
		luckyNumbers[i] = findLucky(numbers)
	}

	// Collect output from finders
	merged := merge(luckyNumbers)
	for n := range merged {
		fmt.Println(n)
	}
}
