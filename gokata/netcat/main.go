// Netcat is a read-only TCP client. Adapted from
// https://github.com/adonovan/gopl.io/tree/master/ch8/netcat1.
package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:1155") // the clock kata listens on this address
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// io.Reader is implemented by conn
	if _, err := io.Copy(os.Stdout, conn); err != nil {
		log.Fatal(err)
	}
}
