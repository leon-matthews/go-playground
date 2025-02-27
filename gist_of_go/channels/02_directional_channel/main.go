// Specify channel direction to catch errors at compile-time
// Channels can be:
//
//	chan (bidirectional)
//	chan<- (send-only)
//	<-chan (receive-only)
package main

import (
	"fmt"
	"strings"
)

func main() {
	str := "one,two,,four"
	stream := make(chan string)
	go submit(str, stream)
	print(stream)
}

// submit can only write to the stream channel
func submit(str string, stream chan<- string) {
	words := strings.Split(str, ",")
	for _, word := range words {
		stream <- word
	}
	close(stream)
}

// print may only read from the stream channel
func print(stream <-chan string) {
	for word := range stream {
		if word != "" {
			fmt.Printf("%s ", word)
		}
	}
	// Closing twice error avoided with a compile-time error:
	// close(stream)
	// cannot close receive-only channel stream (variable of type <-chan string)
	fmt.Println()
}
