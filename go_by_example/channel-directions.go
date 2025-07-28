package main

import (
	"fmt"
)

// You can specify channel direction in function parameters
func main() {
	pings := make(chan string, 1)
	pongs := make(chan string, 1)
	ping(pings, "passed message")
	pong(pings, pongs)
	fmt.Println(<-pongs)
}

// ping can only send to channel
func ping(pings chan<- string, msg string) {
	pings <- msg
}

// pong can only receive over channel
func pong(pings <-chan string, pongs chan<- string) {
	msg := <-pings
	pongs <- msg
}
