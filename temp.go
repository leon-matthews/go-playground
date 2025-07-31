package main

import (
	"fmt"
	"time"
)

func main() {
	var c1, c2 <-chan int // Nil channels always block on read
	start := time.Now()
	select {
	case <-c1:
	case <-c2:
	default:
		fmt.Println("default selected after", time.Since(start))
	}
}
