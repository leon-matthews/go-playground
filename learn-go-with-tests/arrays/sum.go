package main

// Sum returns total of all numbers in given slice
func Sum(numbers []int) int {
	sum := 0
	for _, n := range numbers {
		sum += n
	}
	return sum
}

// SmallAll totals each of the given slices, individually.
func SumAll(slices ...[]int) []int {
	sums := make([]int, 0, len(slices))
	for _, s := range slices {
		sums = append(sums, Sum(s))
	}
	return sums
}

// SumTail totals all but the first element of each of the given slices, individually.
func SumTail(slices ...[]int) []int {
	sums := make([]int, 0, len(slices))
	for _, s := range slices {
		var tail []int
		if len(s) > 0 {
			tail = s[1:]
		}
		sums = append(sums, Sum(tail))
	}
	return sums
}
