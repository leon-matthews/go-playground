package main

import (
	"bytes"
	"fmt"
)

type staffID int

type staff struct {
	id   staffID
	name string
	age  int
}

type manager struct {
	staff
	reports []staffID
}

func main() {
	leon := staff{staffID(3), "Leon", 28}
	eric := staff{2, "Eric", 26}
	alex := manager{
		staff:   staff{1, "Alex", 27},
		reports: []staffID{leon.id, eric.id}}
	fmt.Println("Have:", alex, eric, leon)

	out := bytes.NewBuffer()

}
