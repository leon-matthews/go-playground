package main

import (
	"fmt"
	"time"
)

func main() {
	done := make(chan any)
	terminated := doWork(done, nil)

	// Cancel after one second
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("Cancel doWork()")
		close(done)
	}()

	// Join doWork goroutine
	<-terminated
	fmt.Println("Finished!")
}

// doWork reads strings from the given channel in a separate goroutine
// If done is closed the goroutine will exit and the returned channel will be closed.
func doWork(done <-chan any, strings <-chan string) <-chan any {
	terminated := make(chan any)

	go func() {
		// Don't close terminated until goroutine exits
		defer fmt.Println("doWork exited")
		defer close(terminated)
		for {
			select {
			case s := <-strings:
				fmt.Println(s)
			case <-done:
				// Read only succeeds when channel closed
				return
			}
		}
	}()

	return terminated
}
