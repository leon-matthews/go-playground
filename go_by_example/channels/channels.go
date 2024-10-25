package main

import (
	"fmt"
	"time"
)

func main() {
	unbuffered()
	buffering()
	synchronisation()
	directions()
	selectChannel()
}

// By default sends and receives block until both the sender and
// receiver are ready.
func unbuffered() {
	messages := make(chan string)

	// Send string to channel in a goroutine
	go func() { messages <- "Hello" }()

	msg := <-messages
	fmt.Println(msg)
}

// Here we make a channel of strings buffering up to 2 values.
func buffering() {
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

// We can use channels to synchronize execution across goroutines.
func synchronisation() {
	done := make(chan bool, 1)
	go worker(done)

	// Block until we receive a notification from the worker on the channel
	<-done
}

func worker(done chan bool) {
	fmt.Print("working...")
	time.Sleep(300 * time.Millisecond)
	fmt.Println("done")

	// Send a value to notify that we’re done.
	done <- true
}

// You can specify channel direction in function parameters
func directions() {
	pings := make(chan string, 1)
	pongs := make(chan string, 1)
	ping(pings, "passed message")
	pong(pings, pongs)
	fmt.Println(<-pongs)
}

// Ping can only send to channel
func ping(pings chan<- string, msg string) {
	pings <- msg
}

// Pong can only receive over channel
func pong(pings <-chan string, pongs chan<- string) {
	msg := <-pings
	pongs <- msg
}

// Go's select lets you wait on multiple channel operations.
func selectChannel() {
	c1 := make(chan string)
	c2 := make(chan string)

	go func() {
		time.Sleep(300 * time.Millisecond)
		c1 <- "one"
	}()

	go func() {
		time.Sleep(300 * time.Millisecond)
		c2 <- "two"
	}()

	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-c1:
			fmt.Println("received", msg1)
		case msg2 := <-c2:
			fmt.Println("received", msg2)
		}
	}
}
