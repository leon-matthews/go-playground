package main

import (
	"fmt"
)

func main() {
	generated := rangeGen(41, 46)
	for val := range generated {
		fmt.Println(val)

		// Go routine in rangeGen now permanently blocked trying to send
		// number 43 to the out channel!
		if val == 42 {
			break
		}
	}
}

func rangeGen(start, stop int) <-chan int {
	out := make(chan int)
	go func() {
		for i := start; i < stop; i++ {
			out <- i
		}
		close(out)
	}()
	return out
}
