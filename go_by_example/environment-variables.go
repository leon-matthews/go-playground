package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// Set and get
	os.Setenv("STAGING", "1")
	fmt.Println("STAGING:", os.Getenv("STAGING"))
	fmt.Println("PRODUCTION:", os.Getenv("PRODUCTION")) // empty string

	// Print just the keys from current environment
	fmt.Println()
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		fmt.Println(parts[0])
	}
}
