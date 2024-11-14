package main

import (
	"fmt"
	"math"
	"strings"
)

type Counts map[int]int

// CountLengths builds a map of name lengths vs count
func CountLengths(names []Name) Counts {
	counts := make(Counts)
	for _, name := range names {
		counts[name.Length()]++
	}
	return counts
}

// PrintHistogram groups counts into buckets of binSize and prints basic histogram
func PrintHistogram(counts Counts, binSize int) {
	// Find largest count value
	var largest int
	for k := range counts {
		largest = max(largest, k)
	}

	numBins := int(math.Ceil(float64(largest) / float64(binSize)))
	for bin := 0; bin < numBins; bin++ {
		count := 0
		for i := 0; i < binSize; i++ {
			index := binSize*bin + i
			c, ok := counts[index]
			if ok {
				count += c
			}
		}
		rangeStart := bin * binSize
		rangeEnd := rangeStart + (binSize - 1)

		name := fmt.Sprintf("%2d -> %2d", rangeStart, rangeEnd)
		value := strings.Repeat("#", count)

		fmt.Printf("%s: %s\n", name, value)
	}
}
