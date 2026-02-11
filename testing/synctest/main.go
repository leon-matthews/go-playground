package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("Test synctest")
	trickyOrdering()
}

func trickyOrdering() {
	done := make(chan struct{})
	wait := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	events := make([]string, 0)

	go func() {
		<-wait
		events = append(events, "close done channel")
		close(done)
	}()

	go func() {
		wg.Wait()
		events = append(events, "close wait channel")
		close(wait)
	}()

	go func() {
		time.Sleep(time.Second)
		events = append(events, "finish wait group")
		wg.Done()
	}()

	events = append(events, "blocked on done channel")
	<-done

	for i := range events {
		fmt.Println(i+1, events[i])
	}
}
