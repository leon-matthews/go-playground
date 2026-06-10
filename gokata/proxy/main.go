// Proxy mediates TCP traffic between a client and an upstream server. Adapted from
// http://youtu.be/J4J-A9tcjcA.
package main

import (
	"io"
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		// NOTE: don't put any blocking code here!
		go proxy(conn)
	}
}

func proxy(conn net.Conn) {
	defer conn.Close() // release precious file descriptor

	upstream, err := net.Dial("tcp", "google.com:http")
	if err != nil {
		log.Print(err)
		return
	}

	go io.Copy(upstream, conn) // in this case it's ok not to track the goroutine
	io.Copy(conn, upstream)
}
