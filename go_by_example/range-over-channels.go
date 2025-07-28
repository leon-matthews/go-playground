package main

import (
	"fmt"
)

// Range iterates over each element as it’s received from channel
func main() {
	queue := make(chan string, 5)
	queue <- "one"
	queue <- "two"
	queue <- "three"

	// Without `close()`, program panics
	// fatal error: all goroutines are asleep - deadlock!
	close(queue)

	// Values can still be received from closed channel
	for word := range queue {
		fmt.Println(word)
	}
}
