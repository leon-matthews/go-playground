package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
)

func main() {
	r1 := IncompressableReader(1234567890)
	r2 := IncompressableReader(1234567890)

	buf1 := make([]byte, 512)
	buf2 := make([]byte, 512)

	_, err := r1.Read(buf1)
	if err != nil {
		log.Fatal(err)
	}
	_, err = r2.Read(buf2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("[%T]%+[1]v\n", buf1)
	fmt.Printf("[%T]%+[1]v\n", buf2)
}

// IncompressableReader builds a reader over random, incompressible bytes.
// Every reader created with the same seed will produce the same sequence: the
// nth byte of one will equal the nth byte of the other.
func IncompressableReader(seed int) io.Reader {
	// We're only using a tiny fraction of the real seed's space, as our intent
	// is only to provide several different binary readers for testing.
	s := [32]byte{}
	binary.Encode(s[0:], binary.LittleEndian, uint64(seed))
	return rand.NewChaCha8(s)
}
