package main

import (
	"fmt"
	"time"
)

// We can use channels to synchronize execution across goroutines.
func main() {
	done := make(chan bool)
	go worker(done)

	// Block until we receive a notification from the worker on the channel
	<-done
}

func worker(done chan bool) {
	fmt.Print("working...")
	time.Sleep(300 * time.Millisecond)
	fmt.Println("done")

	// Send a value to notify that we’re done.
	done <- true
}
