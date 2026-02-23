package main

import (
	"fmt"
	"time"
)

func main() {
	messages := make(chan string)

	go func() {
		time.Sleep(1 * time.Second)
		messages <- "ping"
	}()

	// Reading from channel blocks, until message received or channel closed
	fmt.Println(<-messages) // Send and receive are synchronised here
}
