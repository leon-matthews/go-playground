package main

// Add the integers in the given slice
func Sum(numbers []int) int {
	sum := 0
	for _, number := range numbers {
		sum += number
	}
	return sum
}

// Create a sum for every given slice
func SumAll(nested ...[]int) []int {
	var sums []int
	for _, numbers := range nested {
		sums = append(sums, Sum(numbers))
	}

	return sums
}

// Create a sum of all but the first element in ever given slice
func SumAllTails(nested ...[]int) []int {
	var sums []int
	for _, numbers := range nested {
		if len(numbers) == 0 {
			sums = append(sums, 0)
		} else {
			tail := numbers[1:]
			sums = append(sums, Sum(tail))
		}
	}

	return sums
}
