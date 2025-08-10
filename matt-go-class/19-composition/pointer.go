// Struts can also embed a pointer to another type
package main

import (
	"fmt"
)

type Pair struct {
	Path string
	Hash string
}

func (p Pair) String() string {
	return fmt.Sprintf("Hash of %s is %s", p.Path, p.Hash)
}

type Fizgig struct {
	*Pair
	Broken bool
}

func main() {
	fg := Fizgig{
		&Pair{"/home", "0xabcde"},
		false,
	}
	fmt.Println(fg)
}
