package main

import (
	"log"

	"go-playground/monarch/mediainfo"
)

func main() {
	version, err := mediainfo.Version()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Using %s %s", mediainfo.Binary, version)
}
