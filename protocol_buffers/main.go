package main

import (
	"fmt"
	"protocol_buffers/addresses"
)

func main() {
	fmt.Println(addresses.Greeting)
	p := addresses.Person{
		Name:  "Leon Matthews",
		Email: "leon.matthews@example.com",
	}
	fmt.Println(p)
}
