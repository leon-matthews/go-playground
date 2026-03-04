package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	runExample("mergeSequentially()", mergeSequentially)
	runExample("mergeConcurrently()", mergeConcurrently)
	runExample("mergeSelect()", mergeSelect)
}

// mergeSequentially performs no concurrency at all
func mergeSequentially(nums1, nums2 <-chan int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)

		for n := range nums1 {
			out <- n
		}
		for n := range nums2 {
			out <- n
		}
	}()

	return out
}

// mergeConcurrently starts a goroutine for each input channel
func mergeConcurrently(nums1, nums2 <-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	// The first goroutine reads from in1 to out.
	wg.Go(func() {
		for n := range nums1 {
			out <- n
		}
	})

	// The second goroutine reads from in2 to out.
	wg.Go(func() {
		for n := range nums2 {
			out <- n
		}
	})

	// Wait until both input channels are exhausted,
	// then close the output channel.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// mergeSelect merges concurrently using a select statement
func mergeSelect(nums1, nums2 <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)

		// Exit once both channels have been closed
		for numClosed := 0; numClosed < 2; {
			// Set the channels to nil to turn off select, avoiding zero values
			select {
			case n, ok := <-nums1:
				if ok {
					out <- n
				} else {
					nums1 = nil
					numClosed++
				}
			case n, ok := <-nums2:
				if ok {
					out <- n
				} else {
					nums2 = nil
					numClosed++
				}
			}
		}
	}()
	return out
}

// rangeGenerator sends values from start to stop (inclusive) to returned channel
func rangeGenerator(start int, stop int) <-chan int {
	out := make(chan int)
	go func() {
		for i := start; i <= stop; i++ {
			time.Sleep(50 * time.Millisecond)
			out <- i
		}
		close(out)
	}()
	return out
}

func runExample(name string, f func(nums1, nums2 <-chan int) <-chan int) {
	fmt.Println(name)
	start := time.Now()
	nums1 := rangeGenerator(1, 10)
	nums2 := rangeGenerator(10, 20)
	nums := f(nums1, nums2)
	for n := range nums {
		fmt.Print(n, " ")
	}
	fmt.Println()
	fmt.Println(time.Since(start))
	fmt.Println()
}
