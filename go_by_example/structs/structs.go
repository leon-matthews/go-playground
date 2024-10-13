package main

import "fmt"

type Person struct {
	name string
	age  uint
}

func newPerson(name string) *Person {
	p := Person{name: name}
	p.age = 42

	// It looks so VERY wrong, but you *are* allowed to return a pointer
	// to a local variable. Honest! Go is garbage collected!
	return &p
}

func main() {
	// Literal syntax requires all fields
	mum := Person{"Alyson", 43}

	// Omitted fields get their zero-value
	baby := Person{name: "Baby"}

	// Returned from function
	dad := newPerson("Leon")

	fmt.Println(dad, mum, baby)

	// If a struct type is only used for a single value,
	// we don’t have to give it a name.
	cat := struct {
		name string
	}{
		"Old Ben",
	}
	fmt.Println(cat)

	// Anonymous structs are used to build table-driven tests
	dogs := []struct {
		name   string
		isGood bool
	}{
		{"Teddy", true},
		{"Velma", false},
	}
	fmt.Println(dogs)
}
