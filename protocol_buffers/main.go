package main

import (
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"log"
	"protocol_buffers/addresses"
)

func main() {
	fmt.Println(addresses.Greeting)
	person := addresses.Person{
		Name:  "Leon Matthews",
		Email: "leon.matthews@example.com",
		Phones: []*addresses.Person_PhoneNumber{
			{Number: "+64-21-555-4321", Type: addresses.PhoneType_PHONE_TYPE_MOBILE},
		},
	}

	// Binary wire format
	out, err := proto.Marshal(&person)
	if err != nil {
		log.Fatalln("Failed to encode person:", err)
	}
	log.Println("Binary format", len(out), "bytes:")
	fmt.Println(out)

	// Text format
	outText, err := prototext.Marshal(&person)
	if err != nil {
		log.Fatalln("Failed to encode person:", err)
	}
	log.Println("Text format", len(outText), "bytes:")
	fmt.Println(string(outText))

	// Debug text format
	fmt.Println(prototext.Format(&person))

	// Debug text format
	fmt.Println(protojson.Format(&person))
}
