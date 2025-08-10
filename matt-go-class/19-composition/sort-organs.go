package main

import (
	"fmt"
	"sort"
)

type Organ struct {
	Name   string
	Weight int
}

type Organs []Organ

func (s Organs) Len() int {
	return len(s)
}

func (s Organs) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Embedding []Organ into struct brings in methods Len() & Swap()
type ByName struct{ Organs }
type ByWeight struct{ Organs }

// Specialise just the Less() method to change each type's behaviour
func (s ByName) Less(i, j int) bool {
	return s.Organs[i].Name < s.Organs[j].Name
}

func (s ByWeight) Less(i, j int) bool {
	return s.Organs[i].Weight < s.Organs[j].Weight
}

func main() {
	// Organs
	o := []Organ{{"brain", 1340}, {"liver", 1494}, {"spleen", 162}, {"pancreas", 131}, {"heart", 290}}

	sort.Sort(ByName{o})
	fmt.Println(o)

	sort.Sort(ByWeight{o})
	fmt.Println(o)

	// What's similar - and really crazy - is [sort.Reverse]
	// It is just a struct and a less than swaps the arguments of *your* less!
	// type reverse struct { Interface}
	// func (r reverse) Less(i, j int) bool { return r.Interface.Less(j, i) }
	sort.Sort(sort.Reverse(ByWeight{o}))
	fmt.Println(o)
}
