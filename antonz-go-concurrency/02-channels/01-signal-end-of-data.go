package main

import (
	"fmt"
	"strings"
)

// Channel guidelines
//  1. Only the writer can close the channel, not the reader. If the reader
//     closes it, the writer will encounter a panic on the next write.
//  2. A writer can only close the channel if they are the sole owner. If there
//     are multiple writers and one closes the channel, the others will face a
//     panic on their next write or attempt to close the channel.
//  3. You don't have to close a channel for it to be garbage collected. If you
//     don't need to signal to readers that there is no more data, you don't
//     need to close the channel.
func main() {
	str := "one,two,,four"
	split(str)
	// one two four
}

// split prints comma-separated values, skipping empty fields.
func split(str string) {
	in := make(chan string)
	go func() {
		words := strings.Split(str, ",")
		for _, word := range words {
			in <- word
		}
		close(in)
	}()

	for word := range in {
		if word != "" {
			fmt.Printf("%s ", word)
		}
	}
	fmt.Println()
}
