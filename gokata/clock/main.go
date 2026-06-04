// Clock is a TCP server that periodically writes the time. Adapted from
// https://github.com/adonovan/gopl.io/tree/master/ch8/clock2.
package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

const (
	host = "localhost"
	port = "1155"
)

func main() {
	address := net.JoinHostPort(host, port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Clock server listening on %s.\n", address)
	fmt.Printf("(Try running `nc %s %s` from multiple terminals.)\n", host, port)

	for {
		// Wait for the next connection.
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		// Handle each connection in its own goroutine.
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	log.Printf("client %v: connected", conn.RemoteAddr())

	// Delay until the next whole second so ticks line up with the clock.
	time.Sleep(untilNextSecond(time.Now()))
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		_, err := io.WriteString(conn, time.Now().Format("15:04:05\n"))
		if err != nil {
			// The write fails once the client disconnects.
			log.Printf("client %v: %v", conn.RemoteAddr(), err)
			return
		}
		<-ticker.C
	}
}

// untilNextSecond returns the time remaining until the next whole second.
func untilNextSecond(t time.Time) time.Duration {
	return t.Truncate(time.Second).Add(time.Second).Sub(t)
}
