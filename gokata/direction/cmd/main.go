package main

import (
	"fmt"
	"math/rand/v2"

	"github.com/gokatas/direction"
)

func main() {
	// Nothing stops us creating direction.Cardinal(42)
	cardinal := direction.Cardinal(rand.N(4))

	// Will panic if n >= 4
	fmt.Print(cardinal)

	switch cardinal {
	case direction.North:
		fmt.Println(" goes up.")
	case direction.South:
		fmt.Println(" goes down.")
	default:
		fmt.Println(" stays put.")
	}
}
