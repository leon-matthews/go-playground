package main

import (
	"errors"
	"fmt"
)

func main() {
	fmt.Println(div(5, 2))
	simulateNamedParameters()
	variadicFunctions()
	multipleReturnValues()
	functionsAreValues()

	// Closure still has access to
	f := anonymousFunctions()
	f()
}

func div(num int, denom int) int {
	if denom == 0 {
		return 0
	}
	return num / denom
}

type FuncArgs struct {
	FirstName string
	LastName  string
	Age       int
}

func MyFunc(args FuncArgs) {
	// Do something here
}

// Named and optional function parameters can be simulated using a struct
func simulateNamedParameters() {
	MyFunc(FuncArgs{FirstName: "Gandalf", Age: 54_962})
}

func addTo(base int, vals ...int) []int {
	out := make([]int, 0, len(vals))
	for _, v := range vals {
		out = append(out, base+v)
	}
	return out
}

func variadicFunctions() {
	fmt.Println(addTo(10, 1, 2, 3, 4, 5))

	vals := []int{5, 4, 3, 2, 1}
	fmt.Println(addTo(10, vals...))
}

func divAndRemainder(num int, denom int) (int, int, error) {
	if denom == 0 {
		return 0, 0, errors.New("cannot divide by zero")
	}
	return num / denom, num % denom, nil
}

func multipleReturnValues() {
	result, remainder, err := divAndRemainder(5, 2)
	fmt.Println(result, remainder, err)
}

// Two functions with same signature
func f1(a string) int {
	return len(a)
}

func f2(a string) int {
	total := 0
	for _, v := range a {
		total += int(v)
	}
	return total
}

func functionsAreValues() {
	// Note type of variable
	var f func(string) int
	f = f1
	fmt.Println(f("Hello"))
	f = f2
	fmt.Println(f("Hello"))
}

func anonymousFunctions() func() {
	n := "Leon"

	// Anonymous functions can reference outside variables
	func() {
		fmt.Println(n)
	}()

	// And be assigned to a variable
	anon := func() {
		fmt.Println(n)
	}

	// Note that value of `n` at the time of the call is used
	n = "Matthews"
	anon()

	// Return closure
	return anon
}
