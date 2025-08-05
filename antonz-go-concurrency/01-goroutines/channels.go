package main

import (
	"fmt"
)

type counter map[string]int

func main() {
	messages := make(chan string)

	go func() {
		messages <- "ping"
	}()

	fmt.Println(<-messages) // Send and receive are synchronised here
}
