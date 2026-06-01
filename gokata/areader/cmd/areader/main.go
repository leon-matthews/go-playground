package main

import (
	"fmt"

	"github.com/gokatas/areader"
)

func main() {
	data := make([]byte, 3)
	var a areader.A
	a.Read(data) // NOTE: ignoring error since Read never returns one
	fmt.Println(data, string(data))
}
