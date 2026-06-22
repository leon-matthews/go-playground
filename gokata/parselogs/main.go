// Parselogs decodes JSON logs line by line. Adapted from
// https://www.ardanlabs.com/blog/2024/03/for-loops-and-more-in-go.html.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
)

func main() {
	logs := `
	{"timestamp": "2024-06-12", "level": "info", "message": "everything is awesome"}
	{"timestamp": "2024-06-12", "level": "warn", "message": "I have a bad feeling about this"}`
	if err := ParseLogs(strings.NewReader(logs)); err != nil {
		log.Fatal(err)
	}
}

type Log struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

func ParseLogs(r io.Reader) error {
	// A Decoder reads and decodes JSON values from an input stream.
	dec := json.NewDecoder(r)

loop:
	for i := 1; ; i++ {
		// Decode reads the next JSON-encoded value from its input and stores it in the value pointed to by v
		var l Log
		err := dec.Decode(&l)
		switch {
		case errors.Is(err, io.EOF):
			break loop
		case err != nil:
			return fmt.Errorf("decoding JSON: %v", err)
		default:
			fmt.Printf("log %d: %+v\n", i, l)
		}
	}

	return nil
}
