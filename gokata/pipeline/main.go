package main

import "fmt"

// counter generates integers up to (and including) max
func counter(max int) <-chan int {
	out := make(chan int)
	go func() {
		for x := 0; x <= max; x++ {
			out <- x
		}
		close(out)
	}()
	return out
}

// squarer generates the square of the integers from in.
func squarer(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for v := range in {
			out <- v * v
		}
		close(out)
	}()
	return out
}

// printer simply prints every integer in given channel
func printer(in <-chan int) {
	for v := range in {
		fmt.Println(v)
	}
}

func main() {
	naturals := counter(10)
	squares := squarer(naturals)
	printer(squares)
}
