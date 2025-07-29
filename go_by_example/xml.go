package main

import (
	"encoding/xml"
	"fmt"
)

// Plant uses field tags and custom types to represent XML node
// Note in particular the use of [xml.Name] and attr on ID
type Plant struct {
	XMLName xml.Name `xml:"plant"`
	ID      int      `xml:"id,attr"`
	Name    string   `xml:"name"`
	Origin  []string `xml:"origin"`
}

func (p *Plant) String() string {
	return fmt.Sprintf("Plant id=%v, name=%v, origin=%v", p.ID, p.Name, p.Origin)
}

func main() {
	// My favourite plant!
	coffee := &Plant{ID: 27, Name: "Coffee"}
	coffee.Origin = []string{"Ethiopia", "Brazil"}
	fmt.Printf("[%T]%+[1]v\n", coffee)

	// Serialise to XML
	b, _ := xml.MarshalIndent(coffee, " ", "  ")
	fmt.Print(xml.Header) // Generic header
	fmt.Println(string(b))

	// Add list and nesting
	type Nesting struct {
		XMLName xml.Name `xml:"nesting"`
		Plants  []*Plant `xml:"parent>child>plant"`
	}
	tomato := &Plant{ID: 81, Name: "Tomato"}
	tomato.Origin = []string{"Mexico", "California"}

	nesting := &Nesting{}
	nesting.Plants = []*Plant{coffee, tomato}
	b, _ = xml.MarshalIndent(nesting, " ", "  ")
	fmt.Println()
	fmt.Print(xml.Header) // Generic header
	fmt.Println(string(b))
}
