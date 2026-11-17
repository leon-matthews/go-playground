package generics

// Reduce combines all values in c using the function f, begining with start
func Reduce[T any](c []T, f func(T, T) T, start T) T {
	var result = start
	for _, x := range c {
		result = f(result, x)
	}
	return result
}

// Sum calculates the total from a slice of numbers.
func Sum(numbers []int) int {
	add := func(a, b int) int { return a + b }
	return Reduce(numbers, add, 0)
}
