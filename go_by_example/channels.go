package main

import (
	"fmt"
	"time"
)

func main() {
	// Sending and receiving operations are both blocking
	messages := make(chan string)

	// Need to send in a goroutine to avoid deadlock
	go func() {
		// Block here until message received
		time.Sleep(1 * time.Second)
		messages <- "Hello"
	}()

	// Block here until message is available
	msg := <-messages
	fmt.Println(msg)
}
