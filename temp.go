package main

import (
	"fmt"
	"time"
)

func main() {
	go spinner()
	time.Sleep(2 * time.Second)
	fmt.Println("\rFinished!")
}

func spinner() {
	parts := "/-\\|"
	for {
		for _, p := range parts {
			fmt.Printf("\r%c", p)
			time.Sleep(100 * time.Millisecond)
		}
	}
}
