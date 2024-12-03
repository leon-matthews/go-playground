package main

import (
	"fmt"
	"log"
	"os"
	"poker"
)

const dbFilename = "game.db.json"

func main() {
	fmt.Println("Let's play poker")
	fmt.Println("Type '{name} wins' to record a win")

	storage, closer, err := poker.NewFileSystemStorageFromFile(dbFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer closer()

	var dummyAlerter = &poker.BlindAlerterMock{}
	game := poker.NewCLI(storage, os.Stdin, dummyAlerter)
	game.PlayPoker()
}
