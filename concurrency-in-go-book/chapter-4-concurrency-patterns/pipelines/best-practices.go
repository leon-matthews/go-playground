package main

import (
	"fmt"
	"iter"
	"slices"
)

// Every step in the pipeline is now a goroutine and supports cancellation
func main() {
	nums := slices.Collect(count(10))
	done := make(chan any)
	defer close(done)

	out := generator(done, nums...)
	out = multiply(done, out, 2)
	out = add(done, out, 100)

	for i := range out {
		fmt.Println(i)
	}
}

// generator sends given values into the output channel
func generator(done <-chan any, in ...int) <-chan int {
	defer fmt.Println("generator() returned")
	out := make(chan int)
	go func() {
		defer close(out)
		defer fmt.Println("generator()'s goroutine finished")
		for _, i := range in {
			select {
			case out <- i:
			case <-done:
				return
			}
		}
	}()
	return out
}

// add reads from in and writes to returned channel
func add(done <-chan any, in <-chan int, additive int) <-chan int {
	defer fmt.Println("add() returned")
	out := make(chan int)
	go func() {
		defer close(out)
		defer fmt.Println("add()'s goroutine finished")
		for i := range in {
			select {
			case out <- i + additive:
			case <-done:
				return
			}
		}
	}()
	return out
}

// multiply reads from in and writes to returned channel
func multiply(done <-chan any, in <-chan int, multiplier int) <-chan int {
	defer fmt.Println("multiply() returned")
	out := make(chan int)
	go func() {
		defer close(out)
		defer fmt.Println("multiply()'s goroutine finished")
		for i := range in {
			select {
			case out <- i * multiplier:
			case <-done:
				return
			}
		}
	}()
	return out
}

// count returns an iterator over the integers from zero to limit
func count(limit int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := range limit {
			if !yield(i) {
				return
			}
		}
	}
}
