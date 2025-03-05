package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	//stream := make(chan bool)
	stream := make(chan bool, 1) // Now sender doesn't block on send

	send := func() {
		defer wg.Done()
		fmt.Println("sender: ready to send...")
		stream <- true
		fmt.Println("sender: sent!")
	}

	receive := func() bool {
		defer wg.Done()
		fmt.Println("receiver: not ready yet...")
		time.Sleep(1000 * time.Millisecond)
		fmt.Println("receiver: ready to receive...")
		<-stream
		fmt.Println("receiver: received!")
		return true
	}

	go send()
	go receive()
	wg.Wait()
}
