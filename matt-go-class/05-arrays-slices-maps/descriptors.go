package main

import (
	"fmt"
)

func main() {
	t := []byte("string")
	fmt.Println(len(t), t)
	fmt.Println(string(t[len(t)-1]))
}
