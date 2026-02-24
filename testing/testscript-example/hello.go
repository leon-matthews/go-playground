package hello

import (
	"fmt"
	"os"
)

func Main() {
	if len(os.Args[1:]) < 1 {
		fmt.Fprintln(os.Stderr, "usage: hello NAME")
		os.Exit(1)
	}
	fmt.Printf("Hello to you, %s!\n", os.Args[1])
}
