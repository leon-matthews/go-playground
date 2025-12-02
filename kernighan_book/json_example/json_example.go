package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	data := jsonMarshal()
	jsonUnmarshal(data)
}

// Marshal slice of structs to JSON byte string.
func jsonMarshal() []byte {
	type Movie struct {
		Title  string   `json:"title"`
		Year   int      `json:"released"`
		Colour bool     `json:"colour,omitempty"`
		Actors []string `json:"actors"`
	}

	movies := []Movie{
		{"Casablanca", 1942, false, []string{"Bogart", "Bergman"}},
		{"Cool Hand Luke", 1967, true, []string{"Newman"}},
		{"Bullitt", 1968, true, []string{"McQueen", "Bisset"}},
	}

	// Produces '[]byte'
	data, err := json.MarshalIndent(movies, "", "    ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}

	// Print '[]byte' as ASCII string
	fmt.Println("JSON Marshal")
	fmt.Printf("%s\n\n", data)
	return data
}

// Take JSON and populate given struct
func jsonUnmarshal(data []byte) {
	// Anonymous struct holding fields of interest
	var titles []struct {
		Title string
		Year  int `json:"released"`
	}

	if err := json.Unmarshal(data, &titles); err != nil {
		log.Fatalf("JSON unmarshaling failed: %s", err)
	}

	fmt.Println("JSON Unmarshal")
	fmt.Println(titles)
}
