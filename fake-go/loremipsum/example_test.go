package loremipsum_test

import (
	"fmt"
	"math/rand/v2"

	"local.dev/fake-go/loremipsum"
)

// Example generates placeholder text from a seeded source for reproducibility.
func Example() {
	rng := rand.New(rand.NewPCG(42, 42))
	fmt.Println(loremipsum.Words(rng, 5, false))
	fmt.Println(loremipsum.Sentence(rng))
	// Output:
	// dolorum unde quam ullam alias
	// Fugit beatae repudiandae consectetur odio, esse pariatur qui veniam laboriosam architecto impedit fuga sint.
}

// ExampleWords starts with the standard opening words when common is true.
func ExampleWords() {
	rng := rand.New(rand.NewPCG(42, 42))
	fmt.Println(loremipsum.Words(rng, 8, true))
	// Output: lorem ipsum dolor sit amet consectetur adipisicing elit
}
