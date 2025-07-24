package main

import (
	"fmt"
	"time"
)

func main() {
	// Most basic switch statement
	number := Stringify(2)
	fmt.Println(number)

	// Is it the weekend?
	if IsTheWeekend() {
		fmt.Println("It's the weekend!")
	} else {
		fmt.Println("It's a weekday :-(")
	}

	// Switch on type, not value
	GuessType(true)
	GuessType(1)
	GuessType("Hello!")
}

// Highly sophisicated integer to string conversion
func Stringify(number int) (value string) {
	switch number {
	case 1:
		value = "one"
	case 2:
		value = "two"
	case 3:
		value = "three"
	}
	return
}

// Is it the weekend - right at this moment!
func IsTheWeekend() bool {
	// Expression, not variable
	switch time.Now().Weekday() {
	case time.Saturday, time.Sunday:
		// Use multiple expressions in same case statement
		return true

	default:
		// No fall through, explicit default
		return false
	}
}

// Print the type of the given value
func GuessType(i interface{}) {
	// Switch on type, not value
	switch t := i.(type) {
	case bool:
		fmt.Println("That's a boolean you got there")
	case int:
		fmt.Println("You gave me an integer. Thanks, I guess.")

	default:
		fmt.Printf("Don't know what type %T\n", t)
	}
}
