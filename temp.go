package main

import (
	"fmt"
	"math/rand/v2"
)

func main() {
	var seed = [32]byte{0x01}
	source := rand.NewChaCha8(seed)

	// Create 1KiB of random bytes
	b := make([]byte, 1_024)
	source.Read(b)
	fmt.Println(b)

	// Create the exact same bytes
	source.Seed(seed)
	b2 := make([]byte, 1_024)
	source.Read(b2)
	fmt.Println(b2)
}
