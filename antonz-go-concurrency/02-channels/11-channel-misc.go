package main

import "fmt"

func main() {
	closeUnbuffered()
	closeBuffered()
	nilChannel()
}

func closeUnbuffered() {
	stream := make(chan int)
	close(stream)

	// Reading from a closed channel returns zero value and a false status.
	v, ok := <-stream
	fmt.Println(v, ok) // 0, false
}

func closeBuffered() {
	stream := make(chan int, 2)
	stream <- 1
	stream <- 2
	close(stream)

	// As long as there are values in the buffer, the channel returns those
	// values and a true status. Once all values are read, it returns a zero
	// value and a false status, just like a regular channel.
	v, ok := <-stream
	fmt.Println(v, ok) // 1, true
	v, ok = <-stream
	fmt.Println(v, ok) // 2, true
	v, ok = <-stream
	fmt.Println(v, ok) // 0, false
}

func nilChannel() {
	var stream chan int
	fmt.Println(stream) // <nil>

	// A nil channel is an ugly beast:
	// 1. Writing to a nil channel blocks the goroutine forever.
	// 2. Reading from a nil channel blocks the goroutine forever.
	// 3. Closing a nil channel causes a panic.
}
