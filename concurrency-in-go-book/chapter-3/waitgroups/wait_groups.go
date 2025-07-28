package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	simple()
	helloRunner(3)
	helloRunner2(3)
}


// simple is a nice easy [sync.WaitGroup] example
func simple() {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("1st goroutine sleeping")
		time.Sleep(1 * time.Millisecond)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("2nd goroutine sleeping")
		time.Sleep(2 * time.Millisecond)
	}()

	wg.Wait()
	fmt.Println("All goroutines complete")
}

// helloRunner is an improvement: hello2() doesn't know its concurrent
func helloRunner(numGreeters int) {
	wg := sync.WaitGroup{}
	wg.Add(numGreeters)
	for i := range numGreeters {
		go hello(i, &wg)
	}
	wg.Wait()
}

// hello() must take a *pointer* to [wg], not a copy.
func hello(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("hello(%d)\n", id)
}


// helloRunner2 is an improvement: hello2() doesn't know its concurrent
func helloRunner2(numGreeters int) {
	wg := sync.WaitGroup{}
	wg.Add(numGreeters)
	for i := range numGreeters {
		go func() {
			defer wg.Done()
			hello2(i)
		}()
	}
	wg.Wait()
}

// hello2() doesn't need a wait group
func hello2(id int) {
	fmt.Printf("hello2(%d)\n", id)
}
