package main

import (
	"embed"
    "log"

	"etag/experiment/embed_hash"
    "etag/experiment/hasher"
)

//go:embed static/*
var static embed.FS

const name = "static/dbvisit.svg"

func main() {
	hashFS := embed_hash.EmbedHashFS{static}
	f, err := hashFS.Open(name)
	if err != nil {
		log.Fatal(err)
	}

    ch, ok := f.(hasher.FileHasher)
	if !ok {
		log.Fatalf("file does not implement FileHasher: %T", f)
	}
	hash := ch.ContentHash()
	log.Printf("ContentHash(%q) == %s", name, hash)
}
