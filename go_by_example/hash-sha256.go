package main

import (
	"crypto/sha256"
	"fmt"
)

// $ echo -n "Leon Matthews was here" | sha256sum
// 1b9ecf24f67aba3a64878b1f5ad09738c4a3d8d0f648300c36e0470297b08455  -
func main() {
	s := "Leon Matthews was here"
	hasher := sha256.New()  // Interface [hash.Hash] returned, has embedded io.Writer
	hasher.Write([]byte(s)) // Write expects bytes
	bs := hasher.Sum(nil)   // []byte returned
	fmt.Printf("%x\n", bs)  // 1b9ecf24f67aba3a64878b1f5ad09738c4a3d8d0f648300c36e0470297b08455
}
