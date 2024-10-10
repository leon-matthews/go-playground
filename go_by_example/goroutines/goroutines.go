package main

import (
	"fmt"
	"time"
)

func main() {
	// Run function
	f("direct")

	// Then two functions at once
	go f("goroutine")
	go f("goroutine2")

	// Wait for both to finish
	time.Sleep(time.Second)
	fmt.Println("done")
}

// Print a few numbers out
func f(from string) {
	for i := 0; i < 3; i++ {
		fmt.Println(from, ":", i)
		time.Sleep(100 * time.Millisecond)
	}
}
