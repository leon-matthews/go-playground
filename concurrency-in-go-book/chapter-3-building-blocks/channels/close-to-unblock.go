// Close a channel to unblock multiple goroutines at once
package main

import (
	"fmt"
	"sync"
	"time"
)

const numGoroutines = 10

func main() {
	begin := make(chan any)
	wg := sync.WaitGroup{}

	fmt.Println("Creating goroutines")
	for i := range numGoroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Block here trying to read from [begin] channel
			<-begin
			fmt.Printf("%v has begun\n", i)
		}(i)
	}
	fmt.Printf("Created %d goroutines\n", numGoroutines)

	fmt.Print("On your marks... ")
	time.Sleep(time.Second)
	fmt.Print("Get set... ")
	time.Sleep(time.Second)
	fmt.Println("Go!")

	// Once closed, reading will commence (with zero values)
	close(begin)
	wg.Wait()
}
