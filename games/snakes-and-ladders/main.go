package main

import (
	"fmt"
)

func main() {
	var required = 100
	var score int
	for {
		score += D6()
		if score > required {
			break
		}
	}

	fmt.Println("Final score was", score)
}
