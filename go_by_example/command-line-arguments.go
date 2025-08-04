package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("With program: %#v\n", os.Args)
	fmt.Printf("Sans program: %#v\n", os.Args[1:])
}
