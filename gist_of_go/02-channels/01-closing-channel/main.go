// Channels can be closed to communicate that no more values are available.
// The two rules for closing channels are:
//  1. Only the writer can close the channel, not the reader. If the reader
//     closes it, the writer will encounter a panic on the next write.
//  2. A writer can only close the channel if they are the sole owner. If
//     there are multiple writers and one closes the channel, the others will
//     face a panic on their next write or attempt to close the channel.
package main

import (
	"fmt"
	"strings"
)

func main() {
	str := "one,two,,four"

	in := make(chan string)
	go func() {
		words := strings.Split(str, ",")
		for _, word := range words {
			in <- word
		}
		close(in)
	}()

	// For-range statement automatically breaks if channel is closed
	// Note that range over a channel returns a single value, not a
	// pair (unlike range over a slice).
	for word := range in {
		if word != "" {
			fmt.Printf("%s ", word)
		}
	}
}
