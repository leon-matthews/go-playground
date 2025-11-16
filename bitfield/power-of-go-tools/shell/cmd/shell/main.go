package main

import (
	"os"
	"shell"
)

func main() {
	s := shell.NewSession(os.Stdin, os.Stdout, os.Stderr)
	s.Run()
}
