// Basic NATS usage examples.
// Requires that a NATS server is running on the local system, eg.
//
//  $ sudo apt install nats-server
//
// For fun, also run the nats-cli command on a bunch of terminals:
//
//  $ nats subscribe foo

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	fmt.Println("NATS client")
	const url = nats.DefaultURL
	const timeout = 1 * time.Second

	// Connect to server on localhost
	nc, err := nats.Connect(url)
	if err != nil {
		log.Fatalf("could not connect to default NATS server: %v", err)
	}

	// Create Synchronous Subscriber
	sub, err := nc.SubscribeSync("foo")

	// Publish message
	err = nc.Publish("foo", []byte("Hello World"))
	if err != nil {
		log.Fatal("could not publish message: %v", err)
	}

	// Block until message received
	msg, err := sub.NextMsg(timeout)
	if err != nil {
		log.Fatalf("could not receive message: %v", err)
	}
	fmt.Println(string(msg.Data))
}
