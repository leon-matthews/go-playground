package main

import (
	"fmt"
)

// Use select with a default clause to implement non-blocking sends, receives,
// and even non-blocking multi-way selects.
func main() {
	messages := make(chan string)

	// Non-blocking receive
	select {
	case msg := <-messages:
		fmt.Println("received message", msg)
	default:
		fmt.Println("no message received")
	}

	// Non-blocking send
	msg := "Hello"
	select {
	case messages <- msg:
		fmt.Println("sent message")
	default:
		fmt.Println("no message sent")
	}

	// Multi-way non-blocking select
	signals := make(chan bool)
	select {
	case msg := <-messages:
		fmt.Println("received message", msg)
	case sig := <-signals:
		fmt.Println("received signal", sig)
	default:
		fmt.Println("No activity")
	}
}
