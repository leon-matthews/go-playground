// Proxy mediates TCP traffic between a client and an upstream server:
//
//	client <--> proxy (localhost:8000) <--> upstream (google.com:443)
//
// Adapted from http://youtu.be/J4J-A9tcjcA.
package main

import (
	"io"
	"log"
	"net"
)

const (
	// listenAddress is where clients connect to reach the proxy.
	listenAddress = "localhost:8000"

	// upstreamAddress is the real server; "https" resolves to port 443.
	upstreamAddress = "google.com:http"
)

func main() {
	// Listen binds a TCP socket; the kernel then queues incoming connections
	// until we collect them with Accept. Failure (usually the port being in
	// use) is fatal: the proxy is useless without its socket.
	ln, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatal(err)
	}

	// Announce readiness only after Listen succeeds, so the message never lies.
	log.Printf("Forwarding connections on %v to %v", ln.Addr(), upstreamAddress)

	// Accept blocks until a client connects, then returns a net.Conn for just
	// that one client. Handing each connection to its own goroutine frees the
	// loop to Accept the next client immediately.
	for {
		conn, err := ln.Accept()
		if err != nil {
			// Accept errors are usually transient, so log and keep serving.
			log.Print(err)
			continue
		}

		// NOTE: don't put any blocking code here — it would delay every
		// waiting client. The assertion is safe: Accept on a TCP listener
		// always returns a *net.TCPConn, whose CloseWrite method proxy needs.
		go proxy(conn.(*net.TCPConn))
	}
}

// proxy pumps bytes both ways between the client and a new upstream connection.
func proxy(client *net.TCPConn) {
	// Always release the client's file descriptor; leak enough of them and
	// Accept starts failing.
	defer client.Close()

	// Open the second half of the tunnel; Dial covers the DNS and service
	// name lookups. On failure the deferred Close hangs up on the client.
	conn, err := net.Dial("tcp", upstreamAddress)
	if err != nil {
		log.Print(err)
		return
	}
	upstream := conn.(*net.TCPConn) // Dial with "tcp" always returns *TCPConn
	defer upstream.Close()

	// TCP is full-duplex and io.Copy blocks until EOF, so each direction
	// needs its own copy, and hence one more goroutine. CloseWrite sends a
	// TCP FIN — a "half-close": no more data from this side, the other
	// direction stays open. Forwarding each FIN promptly matters: without it
	// a hung-up client goes unnoticed by the upstream. The io.Copy errors are
	// discarded, as peers resetting connections is routine for a proxy.
	done := make(chan struct{})
	go func() {
		// Client -> upstream; on EOF, pass the client's FIN along.
		io.Copy(upstream, client)
		upstream.CloseWrite()
		close(done)
	}()

	// Upstream -> client runs here, so proxy lives for as long as the
	// upstream keeps talking.
	io.Copy(client, upstream)
	client.CloseWrite()

	// Wait for the other direction to finish before the deferred Closes tear
	// both sockets down — the client may still be sending.
	<-done
}
