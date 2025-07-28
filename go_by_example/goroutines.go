// A goroutine is a lightweight thread of execution.
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	waitSleep()
	waitWaitGroup()
}

// Sleep main thread for a 'while' so hopefully goroutines finish
func waitSleep() {
	// Invoke function directly
	f("direct")

	// Invoke function in goroutine
	go f("goroutine")

	// Run anonymous function in goroutine
	go func(message string) {
		fmt.Println(message)
	}("anonymous")

	// No output from goroutines unless we wait
	time.Sleep(10 * time.Millisecond)
	fmt.Println("finished")
}

// Use [sync.WaitGroup] to ensure that all goroutines have finished
func waitWaitGroup() {
	var wg sync.WaitGroup

	inner := func(message string) {
		defer wg.Done()
		for i := 0; i < 3; i++ {
			fmt.Println(message, ":", i)
		}
	}

	wg.Add(1)
	go inner("goroutine #1")

	wg.Add(1)
	go inner("goroutine #2")

	wg.Wait()
}

func f(from string) {
	for i := 0; i < 3; i++ {
		fmt.Println(from, ":", i)
	}
}
