package main

import (
	"sync"
	"testing"
)

func main() {

}


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
