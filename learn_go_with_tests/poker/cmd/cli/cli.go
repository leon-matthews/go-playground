package main

import (
	"fmt"
	"learn_go_with_tests/poker"
	"log"
	"os"
)

const filename = "poker.db"

func main() {
	fmt.Println("Let's play poker")
	fmt.Println("Type '{name} wins' to record a win")
	store, err := poker.NewPlayerStoreBolt(filename)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	game := poker.NewCLI(store, os.Stdin, nil)
	game.PlayPoker()
}
