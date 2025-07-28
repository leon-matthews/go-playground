package main

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

type readOp struct {
	key  int
	resp chan int
}

type writeOp struct {
	key  int
	val  int
	resp chan bool
}

func main() {
	const (
		numReaders = 100
		numWriters = 10
	)
	var (
		numReads, numWrites uint64
	)

	// Channel for requests TO the 'stateful' goroutine, from its clients.
	reads := make(chan readOp)
	writes := make(chan writeOp)

	// Goroutine has private state, to guarantee it won't get corrupted.
	// It listens for messages from other goroutines to either read or write
	// to that state, via different channels and message types
	go func() {
		// Initialise state
		var state = make(map[int]int)

		// Loop forever over read and write channels
		// Send values back over separate channels in message type
		for {
			select {
			case read := <-reads:
				read.resp <- state[read.key]
			case write := <-writes:
				state[write.key] = write.val
				write.resp <- true
			}
		}

		// We're never going to exit!
	}()

	// Start 100 readers...
	for range numReaders {
		go func() {
			for {
				// Create outgoing request
				// (Kind of mad to create a new channel for the response in the inner loop!)
				read := readOp{
					key:  rand.Intn(5),
					resp: make(chan int),
				}
				reads <- read
				<-read.resp
				atomic.AddUint64(&numReads, 1)
				time.Sleep(time.Millisecond)
			}
		}()
	}

	// ...and 10 writers
	for range numWriters {
		go func() {
			for {
				write := writeOp{
					key:  rand.Intn(5),
					val:  rand.Intn(100),
					resp: make(chan bool),
				}
				writes <- write
				<-write.resp
				atomic.AddUint64(&numWrites, 1)
				time.Sleep(time.Millisecond)
			}
		}()
	}

	// Run all 101 goroutines for one second
	time.Sleep(time.Second)

	readOpsFinal := atomic.LoadUint64(&numReads)
	fmt.Println("readOps:", readOpsFinal)
	writeOpsFinal := atomic.LoadUint64(&numWrites)
	fmt.Println("writeOps:", writeOpsFinal)

}
