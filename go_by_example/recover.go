package main

import (
	"fmt"
)

func willPanic() {
	panic("I'm so stuck right now!")
}

func main() {
	fmt.Println("Before panicking")

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered:", r)
		}
	}()

	willPanic()

	// We never see this as we can only recover in `defer()`
	fmt.Println("After panicking")
}
