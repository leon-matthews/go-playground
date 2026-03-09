package main

import (
	"fmt"
	"math/rand/v2"
	"time"
)

func main() {
	var average, largest, smallest time.Duration
	for {
		j := jitter()
		largest = max(largest, j)
		smallest = min(smallest, j)
		average = emv(j, average, 0.1)
		fmt.Printf("\ravg: %6v, min: %6v\tmax: %6v\t", average, smallest, largest)
	}
}

// emv calculates a rough moving average for the given durations
func emv(current, previous time.Duration, alpha float64) time.Duration {
	output := alpha*float64(current) + (1.0-alpha)*float64(previous)
	return time.Duration(output)
}

// jitter measures the difference between the requested and actual delay for a random timer duration
func jitter() time.Duration {
	expected := rand.N(100 * time.Millisecond)
	timer := time.NewTimer(expected)
	start := time.Now()
	<-timer.C
	actual := time.Since(start)
	delta := actual - expected
	return delta
}
