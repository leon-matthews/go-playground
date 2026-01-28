package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
)

type staffID int

type staff struct {
	ID   staffID
	Name string
}

type manager struct {
	staff
	Reports []staffID
}

func main() {
	// Create data
	leon := staff{staffID(7), "Leon"}
	eric := staff{6, "Eric"}
	yang := staff{5, "Yang"}
	alex := manager{
		staff:   staff{1, "Alex"},
		Reports: []staffID{leon.ID, eric.ID, yang.ID},
	}
	fmt.Println("Have:", alex)

	// Encode to GOB
	out := &bytes.Buffer{}
	enc := gob.NewEncoder(out)
	err := enc.Encode(alex)
	if err != nil {
		log.Fatal("Encoding GOB:", err)
	}
	fmt.Println("Encoded:", out.Bytes())

	// Decode from GOB
	dec := gob.NewDecoder(out)
	var alex2 staff
	err = dec.Decode(&alex2)
	if err != nil {
		log.Fatal("Decoding GOB:", err)
	}
	fmt.Println("Decoded:", alex2)
}
