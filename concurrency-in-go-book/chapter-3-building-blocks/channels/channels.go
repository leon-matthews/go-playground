// Channels are synchronisation primitives from Hoare's CSP
// They are best uqsed to communicate between goroutines. Values stream from
// where they are put in to where they come out.
package main

import (
	"fmt"
)

func main() {
	greetings := make(chan string)

	// Send one value then close channel
	go func() {
		defer close(greetings)
		greetings <- "Hello"
	}()

	// Read, printing if 'real' or not
	greeting, ok := <-greetings
	fmt.Printf("(%v): %s\n", ok, greeting)

	// Reading from closed channel - not 'real'
	greeting, ok = <-greetings
	fmt.Printf("(%v): %s\n", ok, greeting)
}
