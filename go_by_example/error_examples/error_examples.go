// Examples of returning errors from functions
package main

import (
	"errors"
	"fmt"
)

func main() {
	// Simple example
	for _, i := range []int{40, 41, 42, 43} {
		if r, e := f(i); e != nil {
			fmt.Println("f failed:", e)
		} else {
			fmt.Println("f worked:", i, "->", r)
		}
	}

	// Making tea
	for i := range 5 {
		if err := makeTea(i); err != nil {
			if errors.Is(err, ErrOutOfTea) {
				fmt.Println("We should buy more tea!")
			} else if errors.Is(err, ErrPower) {
				fmt.Println(err)                // making tea: can't boil water
				fmt.Println(errors.Unwrap(err)) // can't boil water
				fmt.Println("Now it's dark.")
			} else {
				fmt.Printf("Unknown error: %s\n", err)
			}

			continue
		}

		fmt.Println("Tea is ready")
	}
}

// By convention, errors are returned last and have interface type `error`:
//
//	type error interface {
//		Error() string
//	}
//
// The function [errors.New()] builds an error of type [errors.errorString], a
// structure containing just a single string value.
func f(num int) (int, error) {
	if num == 42 {
		return -1, errors.New("can't work with 42")
	}

	return num + 3, nil
}

// A sentinel error is a predeclared variable that is used to
// signify a specific error condition.
// [fmt.Errorf] Builds instance of [errors.errorString]
var ErrOutOfTea = fmt.Errorf("no more tea available")
var ErrPower = fmt.Errorf("can't boil water")

func makeTea(arg int) error {
	if arg == 2 {
		return ErrOutOfTea
	} else if arg == 4 {
		return fmt.Errorf("making tea: %w", ErrPower)
	}
	return nil
}
