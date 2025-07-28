package main

import (
	"fmt"
	"time"
)

// Go's select lets you wait on multiple channel operations.
func main() {
	c1 := make(chan string)
	c2 := make(chan string)

	go func() {
		time.Sleep(300 * time.Millisecond)
		c1 <- "one"
	}()

	go func() {
		time.Sleep(200 * time.Millisecond)
		c2 <- "two"
	}()

	// Running select with no other goroutines running causes a panic:
	// fatal error: all goroutines are asleep - deadlock!
	for range 2 {
		select {
		case msg1 := <-c1:
			fmt.Println("received", msg1)
		case msg2 := <-c2:
			fmt.Println("received", msg2)
		}
	}
}
