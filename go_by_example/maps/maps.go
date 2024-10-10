package main

import (
	"fmt"
	"maps"
)

func main() {
	CreateMaps()
	SetAndGet()
	DeleteAndClear()
	Compare()
}

func CreateMaps() {
	// Using built-in make()
	m := make(map[string]int)
	fmt.Println(len(m))
	m["zero"] = 0
	m["one"] = 1
	fmt.Println(m)

	// Declare `nil` map
	var m2 map[string]int
	fmt.Println(m2 == nil)

	// Non-nill empty
	m3 := map[string]int{}
	fmt.Println(m3 == nil)

	// Declare and initialise
	m4 := map[string]int{"two": 2, "three": 3}
	fmt.Println(m4)
}

func SetAndGet() {
	m := make(map[int]string)
	m[0] = "zero"
	m[1] = "one"
	m[2] = "two"

	// Fetching missing key evaluates to value's zero value
	v := m[100]

	// Or the zero value plus a boolean
	v, ok := m[200]
	fmt.Println("Found", v, ok, "len: ", len(m))
}

func DeleteAndClear() {
	m := map[string]int{"one": 1, "two": 2, "three": 3}
	fmt.Println("Before:", m, "len", len(m))

	// Built-in `delete()` function
	delete(m, "one")
	delete(m, "one") // It doesn't mind missing keys
	delete(m, "zero")
	fmt.Println("After delete():", m, "len", len(m))

	// Remove all entries with `clear()`
	clear(m)
	fmt.Println("After clear():", m, "len", len(m))
}

func Compare() {
	// maps.Equal() compares maps, insertion order not important
	m := map[string]int{"one": 1, "two": 2, "three": 3}
	m2 := map[string]int{"three": 3, "two": 2, "one": 1}
	equal := maps.Equal(m, m2)
	fmt.Println("Maps are the same?", equal)
}
