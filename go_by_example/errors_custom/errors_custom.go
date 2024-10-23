// Custom implementation of error interface
package main

import (
	"errors"
	"fmt"
)

func main() {
	_, err := f(42)
	fmt.Println(err) // "42 - can't work with it"

	var ae *argError
	if errors.As(err, &ae) {
		fmt.Println(ae.arg)     // "42"
		fmt.Println(ae.message) // "can't work with it"
	} else {
		fmt.Println("err doesn't match ArgError")
	}
}

// Return our custom error
func f(arg int) (int, error) {
	if arg == 42 {
		return -1, &argError{arg, "can't work with it"}
	}
	return arg + 3, nil
}

// A custom error type usually has the suffix 'Error'
type argError struct {
	arg     int
	message string
}

// Implement the error interface
func (e *argError) Error() string {
	return fmt.Sprintf("%d - %s", e.arg, e.message)
}
