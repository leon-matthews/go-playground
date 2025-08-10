package main

import (
	"fmt"
	"path/filepath"
)

type Pair struct {
	Path string
	Hash string
}

func (p Pair) String() string {
	return fmt.Sprintf("Hash of %s is %s", p.Path, p.Hash)
}

// Embedded struct's fields promoted to the level of the embedding struct
type PairWithLength struct {
	Pair
	Length int
}

// We can use Pair's String method in our own string method
func (p PairWithLength) String() string {
	return fmt.Sprintf("%s, length %d", p.Pair, p.Length)
}

// Filename takes a Pair. Can it handle a PairWithLength?
// No. A PairWithLength *has* a Pair, but not is-a Pair
func (p Pair) Filename() string {
	return filepath.Base(p.Path)
}

// Interface type
type Filenamer interface {
	Filename() string
}

func main() {
	// Pair
	p := Pair{"/usr", "0xfdfe"}
	fmt.Println(p)
	fmt.Println(p.Filename())

	// PairWithLength
	pl := PairWithLength{Pair{"/home", "0xcafe"}, 13}
	fmt.Println(pl, pl.Length)
	fmt.Println(pl.Filename())

	// Filenamer Interface
	var fn Filenamer = PairWithLength{Pair{"/bin", "0xdead"}, 17}
	fmt.Println(fn.Filename())
}
