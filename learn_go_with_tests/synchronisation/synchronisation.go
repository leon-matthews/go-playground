package main

import "fmt"

func main() {
	fmt.Println("Synchronisation Example")
}

type Counter struct{}

func (c *Counter) Increment() {
	fmt.Println("Incrementing")
}

func (c *Counter) Value() int {
	return 3
}
