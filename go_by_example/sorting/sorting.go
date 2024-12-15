package main

import (
	"cmp"
	"fmt"
	"slices"
)

func main() {
	sortSlices()
	sortByLength()
	sortByField()
}

func sortSlices() {
	strs := []string{"c", "a", "b"}
	slices.Sort(strs)
	fmt.Println("Strings:", strs)

	ints := []int{7, 2, 4}
	slices.Sort(ints)
	fmt.Println("Ints:   ", ints)

	s := slices.IsSorted(ints)
	fmt.Println("Sorted: ", s)
}

// sortByLength sorts strings by their length
func sortByLength() {
	fruit := []string{"peach", "banana", "kiwi"}
	lenCmp := func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	}
	slices.SortFunc(fruit, lenCmp)
	fmt.Println(fruit)
}

// sortByField extracts field for sorting
func sortByField() {
	type Person struct {
		name string
		age  int
	}

	people := []Person{
		Person{name: "Jax", age: 37},
		Person{name: "TJ", age: 25},
		Person{name: "Alex", age: 72},
	}

	compareAges := func(a, b Person) int {
		return cmp.Compare(a.age, b.age)
	}

	slices.SortStableFunc(people, compareAges)
	fmt.Println(people)
}
