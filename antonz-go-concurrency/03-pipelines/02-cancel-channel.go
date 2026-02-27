package main

import (
	"fmt"
)

func main() {
	// Workers should quit when cancel is closed
	cancel := make(chan struct{})
	defer close(cancel)

	generated := rangeGen(cancel, 10, 20)
	for val := range generated {
		fmt.Println(val)
		if val == 15 {
			break
		}
	}
}

func rangeGen(cancel <-chan struct{}, start, stop int) <-chan int {
	out := make(chan int)
	go func() {
		for i := start; i < stop; i++ {
			select {
			case <-cancel:
				fmt.Println("Cancelled")
				return
			case out <- i:
			}
		}
		fmt.Println("close out")
		close(out)
	}()
	return out
}
