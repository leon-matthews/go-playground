package main

import (
	"sync"
	"testing"
)

func main() {

}

// Linux 6.14.0
// AMD Ryzen 9 5950X (16-cores, 32-threads)
//
// $ taskset -c 0 perf bench sched pipe -T
// # Running 'sched/pipe' benchmark:
// # Executed 1000000 pipe operations between two threads
// Total time: 5.315 [sec]
// 5.315496 usecs/op
// 188129 ops/sec
//
// 5.3us to send and receive message on a thread
// 2.6us per context switch

func BenchmarkContextSwitch(b *testing.B) {
	var wg sync.WaitGroup
	begin := make(chan struct{})
	c := make(chan struct{})

	// Use 'zero sized' empty struct as token to pass around
	var token struct{}

	sender := func() {
		defer wg.Done()

		// Wait until signalled to begin
		<-begin
		for i := 0; i < b.N; i++ {
			// Send token to receiver
			c <- token
		}
	}

	receiver := func() {
		defer wg.Done()
		<-begin
		for i := 0; i < b.N; i++ {
			// Recieve token from sender
			<-c
		}
	}

	wg.Add(2)
	go sender()
	go receiver()
	b.StartTimer()
	close(begin)
	wg.Wait()
}
