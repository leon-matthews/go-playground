// Package basics works through JSON basic examples from Go blog
// https://go.dev/blog/json
package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	// Create JSON
	m := Message{"Alice", "Hello", 1294706395881547000}
	b := encode(m)
	fmt.Println(string(b))

	b2 := []byte(`{"Name":"Alice","Body":"Hello","Time":1294706395881547000}`)
	m2 := decode(b2)
	fmt.Println(m2)
}

type Message struct {
	Name string
	Body string
	Time int64
}

// decode initialises struct from JSON string using Unmarshal
func decode(b []byte) Message {
	var m Message
	err := json.Unmarshal(b, &m)
	if err != nil {
		panic(err)
	}
	return m
}

// encode creates JSON string using Marshal to convert struct to []byte
func encode(m Message) []byte {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b // `{"Name":"Alice","Body":"Hello","Time":1294706395881547000}`
}
