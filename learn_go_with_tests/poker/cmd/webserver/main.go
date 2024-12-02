package main

import (
	"log"
	"net/http"
	"poker"
)

const dbFilename = "poker.db.json"
const serverAddress = ":8080"

func main() {
	storage, closer, err := poker.NewFileSystemStorageFromFile(dbFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer closer()
	server := poker.NewPlayerServer(storage)
	log.Println("starting server on", serverAddress)
	if err := http.ListenAndServe(serverAddress, server); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
