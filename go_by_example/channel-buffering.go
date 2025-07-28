package main

import (
	"fmt"
)

func main() {
	// Here we make a channel of strings buffering up to 2 values.
	messages := make(chan string, 2)

	messages <- "buffered"
	messages <- "buffered2"
	// Too may writes cause a panic!
	// fatal error: all goroutines are asleep - deadlock!
	//~ messages <- "buffered3"

	fmt.Println(<-messages)
	fmt.Println(<-messages)

	// Too many reads cause panic
	// fatal error: all goroutines are asleep - deadlock!
	//~ fmt.Println(<- messages)
}
