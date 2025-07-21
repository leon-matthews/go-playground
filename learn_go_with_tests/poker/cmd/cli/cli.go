// Command-line poker client
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
	alerter := poker.AlerterFunc(poker.StdOutAlerter)
	game := poker.NewCLI(store, os.Stdin, os.Stdout, alerter)
	err = game.PlayPoker()
	if err != nil {
		log.Fatalf("Error %v", err)
	}
}
