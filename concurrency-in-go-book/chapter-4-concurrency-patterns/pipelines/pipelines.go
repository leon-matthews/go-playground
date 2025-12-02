package main

import (
	"fmt"
)

// Batch processing, no concurrency
func main() {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Println(nums)

	nums = multiply(nums, 2)
	fmt.Println(nums)

	nums = add(nums, 2)
	fmt.Println(nums)

	nums = multiply(nums, 3)
	fmt.Println(nums)
}

func add(values []int, additive int) []int {
	output := make([]int, len(values))
	for i, v := range values {
		output[i] = v + additive
	}
	return output
}

func multiply(values []int, multiplier int) []int {
	output := make([]int, len(values))
	for i, v := range values {
		output[i] = v * multiplier
	}
	return output
}
