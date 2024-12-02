package main

import (
	"log"
	"net/http"
	"os"
)

const dbFilename = "poker.db.json"
const serverAddress = ":8080"

func main() {
	db, err := os.OpenFile(dbFilename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("could not open database file: %v: %v", dbFilename, err)
	}
	storage, err := NewFileSystemStorage(db)
	if err != nil {
		log.Fatalf("problem creating file system player store, %v ", err)
	}

	server := NewPlayerServer(storage)
	log.Println("starting server on", serverAddress)
	if err := http.ListenAndServe(serverAddress, server); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
