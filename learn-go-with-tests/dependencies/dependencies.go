package main

import (
	"fmt"
	"io"
	"os"
)

func Greet(out io.Writer, name string) {
	fmt.Fprintf(out, "Hello, %s!", name)
}

func main() {
	Greet(os.Stdout, "Leon")
}
