package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func main() {
	// Create order with nested items
	original := fakeOrder()
	fmt.Println(original)

	// Marshal into JSON
	data, err := json.Marshal(original)
	if err != nil {
		log.Fatal("could not marshal object")
	} else {
		fmt.Printf("marshaled into %d bytes\n", len(data))
	}

	// Round trip back to struct
	var order Order
	err = json.Unmarshal(data, &order)
	if err != nil {
		log.Fatal("could not unmarshal data")
	}
	fmt.Println(order)
}

type Order struct {
	ID          string    `json:"id"`
	DateOrdered time.Time `json:"date_ordered"`
	CustomerID  string    `json:"customer_id"`
	Items       []Item    `json:"items"`
}

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Creates a fake order for testing
func fakeOrder() *Order {
	items := []Item{
		Item{"F32", "Size Seven Shoes"},
		Item{"F33", "Size Eight Shoes"},
	}
	o := Order{
		ID:          "54554",
		DateOrdered: time.Now(),
		CustomerID:  "leon@example.com",
		Items:       items,
	}
	return &o
}
