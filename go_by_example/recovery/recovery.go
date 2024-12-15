package main

import "fmt"

func main() {
	defer recovery()
	willPanic()
}

func willPanic() {
	panic("I told you I'd panic!")
}

func recovery() {
	if r := recover(); r != nil {
		fmt.Println("Recovered:", r)
	}
}
