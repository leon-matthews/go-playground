package main

import "fmt"

func main() {
	cancel := make(chan struct{})
	defer close(cancel)

	generated := rangeGen(cancel, 41, 46)
	for val := range generated {
		fmt.Println(val)

		// Go routine in rangeGen now permanently blocked trying to send
		// number 43 to the out channel!
		if val == 42 {
			cancel <- struct{}{}
			break
		}
	}
}

func rangeGen(cancel chan struct{}, start, stop int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := start; i < stop; i++ {
			// If cancel is closed, exit the goroutine;
			// Otherwise, send the next value to out.
			select {
			case out <- i:
			case <-cancel:
				return
			}
		}
	}()
	return out
}
