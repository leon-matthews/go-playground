package main

import (
	"fmt"
	"os"
)

func main() {
	// Temporary file with suffix
	t, err := os.CreateTemp("", "go-by-example-*.txt")
	if err != nil {
		panic(err)
	}
	fmt.Println("Created:", t.Name())

	// Unnecessarily verbose deleting
	defer func() {
		os.Remove(t.Name())
		fmt.Println("Deleted:", t.Name())
	}()

	// Temporary folder example in `directories.go`
}
