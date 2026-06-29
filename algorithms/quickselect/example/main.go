// Command quickselect demonstrates the quickselect package by finding order
// statistics of a shuffled slice of integers without fully sorting it.
package main

import (
	"cmp"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"

	"local.dev/quickselect"
)

func main() {
	count := flag.Int("count", 15, "size of the slice to select from (values 1..count)")
	k := flag.Int("k", 7, "which order statistic to find (0-based)")
	flag.Parse()

	if *count < 1 {
		log.Fatal("count must be at least 1")
	}
	if *k < 0 || *k >= *count {
		log.Fatalf("k=%d out of range [0, %d)", *k, *count)
	}

	// NthElement reorders the slice in place, so print it before selecting.
	numbers := shuffledIntegers(*count)
	fmt.Println("input:      ", numbers)

	nth := quickselect.NthElement(numbers, *k)
	fmt.Printf("the k=%d smallest element is %d\n\n", *k, nth)

	// values[k] now sits in its sorted position, with everything smaller to its
	// left and everything larger to its right. Each side is itself still
	// unordered -- that is the sorting work quickselect avoids.
	fmt.Println("partitioned:", numbers)
	fmt.Println("  left  (<=):", numbers[:*k])
	fmt.Println("  nth       :", numbers[*k])
	fmt.Println("  right (>=):", numbers[*k+1:])

	// NthElementFunc takes a comparator. A descending one reverses the order,
	// so the "0th smallest" is really the largest element.
	descending := func(a, b int) int { return cmp.Compare(b, a) }
	largest := quickselect.NthElementFunc(shuffledIntegers(*count), 0, descending)
	fmt.Printf("\nwith a descending comparator, k=0 yields the largest: %d\n", largest)
}

// shuffledIntegers returns the values 1..count in random order.
func shuffledIntegers(count int) []int {
	numbers := make([]int, count)
	for i := range numbers {
		numbers[i] = i + 1
	}
	rand.Shuffle(count, func(i, j int) {
		numbers[i], numbers[j] = numbers[j], numbers[i]
	})
	return numbers
}
