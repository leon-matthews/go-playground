package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	senderWaits()
	bufferedChannel()
}

// Our sender has to wait until the receiver is ready.
func senderWaits() {
	stream := make(chan int)

	send := func() {
		fmt.Println("Sender: ready")
		v := 42
		stream <- v
		fmt.Println("Sender: sent", v)
	}

	receive := func() {
		fmt.Println("Receiver: not ready yet")
		time.Sleep(1 * time.Second)
		fmt.Println("Receiver: ready")
		var v = <-stream
		fmt.Println("Receiver: received", v)
	}

	var wg sync.WaitGroup
	wg.Go(receive)
	wg.Go(send)
	wg.Wait()
}

// A buffered channel allows us to keep writing
func bufferedChannel() {
	stream := make(chan int, 3)
	stream <- 1           // 1 [] []
	stream <- 2           // 1  2 []
	stream <- 3           // 1  2  3
	fmt.Println(<-stream) // 2  3 []
	fmt.Println(<-stream) // 3 [] []

	stream <- 4 // 3  4 []
	stream <- 5 // 3  4  5
	// Would block:
	// stream <- 6
	fmt.Println("Stream length:", len(stream), "capacity:", cap(stream), "free capacity:", len(stream) - cap(stream))
}
