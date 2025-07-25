package main

import "fmt"

func main() {
	basics()
	mulipleReturnValues()
	variadicFunctions()
	closures()
	recursion()
}

func basics() {
	fmt.Println(plus(1, 22))
	fmt.Println(plusPlus(1, 22, 333))
}

func plus(a int, b int) int {
	return a + b
}

func plusPlus(a, b, c int) int {
	return a + b + c
}

func mulipleReturnValues() {
	a, b := vals()
	fmt.Println(a, b)
}

func vals() (int, int) {
	return 3, 4
}

func variadicFunctions() {
	// Pass arguments in manually
	sums(1, 22, 333, 4444, 55555)

	// Or pass in a slice using `...` operator
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sums(nums...)
}

func sums(nums ...int) int {
	fmt.Print(nums, " ")
	total := 0
	for _, num := range nums {
		total += num
	}
	fmt.Println(total)
	return total
}

// Function returns a closure that itself returns an int
func intSeq(start int) func() int {
	i := start
	return func() int {					// Anonymous function
		i++
		return i
	}
}

func closures() {
	nextInt := intSeq(4)
	fmt.Print("Closure: ")
	fmt.Print(nextInt())
	fmt.Print(nextInt())
	fmt.Print(nextInt())
	fmt.Println()
}

func factorial(n uint) uint {
	if n == 0 {
		return 1
	}
	return n * factorial(n-1)
}

func recursion() {
	fmt.Println("factorial(50):", factorial(50))

	var fib func(n int) int

	fib = func(n int) int {
		if n < 2 {
			return n
		}
		return fib(n-1) + fib(n-2)
	}


	fmt.Println("fibonacci(40)", fib(40))
}
