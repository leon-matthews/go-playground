// Package structureless experiments with unmarshalling and modifying
// JSON data without knowing its whole structure in advance - but while
// preserving unused fields when marshaling back to JSON.
package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	b := []byte(`{"Name":"Wednesday","Age":6,"Parents":{"Father": "Gomez", "Mother": "Morticia"},"Siblings": ["Pugsley"]}`)

	// Unmarshall byte slice to any
	var child any
	err := json.Unmarshal(b, &child)
	if err != nil {
		panic(err)
	}

	changeAge(child, 7)
	changeFather(child, "Uncle Fester")

	out, err := json.MarshalIndent(child, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}

// changeAge changes child's age
func changeAge(a any, age int) {
	root, ok := a.(map[string]any)
	if !ok {
		panic("json root not a map")
	}
	root["Age"] = age
}

// changeFather replaces child's father with their real father
func changeFather(a any, realFather string) {
	root, ok := a.(map[string]any)
	if !ok {
		panic("json root not a map")
	}

	parents, ok := root["Parents"].(map[string]any)
	if !ok {
		panic("parents not a map")
	}

	parents["Father"] = realFather
}
