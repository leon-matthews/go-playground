package main

import (
	"fmt"
	"os"
)

// Use os.Exit to immediately exit with a given status.
func main() {
	fmt.Println("Exit code 3 example")
	defer fmt.Println("Exited normally?")
	os.Exit(3)
}
