package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	var n uint64
	var sum uint64
	for range 1_000_000_000 {
		n++
		sum += n
	}
	fmt.Println(time.Since(start))	// time.Now().Sub(start)
	fmt.Println(sum)
}
