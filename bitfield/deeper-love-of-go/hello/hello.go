package main

import (
    "fmt"
    "time"
)

var message = "Hello"

func B() {
    message = "Goodbye"
	for i := range 10 {
		fmt.Println(message, "from goroutine B!", i)
		time.Sleep(50 * time.Millisecond)
	}
}

func main() {
    go B()
	for i := range 10 {
		fmt.Println(message, "from goroutine A!", i)
		time.Sleep(50 * time.Millisecond)
	}
}
