package main

import "fmt"

func main() {
	// Integer /////////////////////////////////////////////
	for i := range 10 {
		fmt.Print(i)
	}
	fmt.Println()

	// Arrays & Slices /////////////////////////////////////

	primes := []int{2, 3, 5, 7}

	// ...ignoring indices
	for _, p := range primes {
		fmt.Print(p)
	}
	fmt.Println()

	// ...index from zero
	for index, p := range primes {
		fmt.Println(index, ":", p)
	}

	// Maps ////////////////////////////////////////////////
	greenstuffs := map[string]string{
		"a": "apple",
		"b": "banana",
		"c": "carrot",
	}

	// ...can iterate over keys
	for k := range greenstuffs {
		fmt.Print(k)
	}
	fmt.Println()

	// ...or over key/value pairs
	for k, v := range greenstuffs {
		fmt.Println(k, v)
	}

	// Strings /////////////////////////////////////////////
	s := "Go ʕ◔ϖ◔ʔ"
	for i, c := range s {
		fmt.Println(i, c)
	}
}
