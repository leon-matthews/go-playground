package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

const address = "127.0.0.1:8090"

func main() {
	// Start TCP server
	log.Println("Listening on:", address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	// Loop forever, start a new goroutine for each incoming connection
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// Housekeeping
	remote := conn.RemoteAddr().String()
	defer func() {
		log.Printf("%s Disconnect", remote)
		conn.Close()
	}()
	conn.SetDeadline(time.Now().Add(time.Second))
	log.Printf("%s Connected", remote)

	// Read line from client
	reader := bufio.NewReader(conn)
	message, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("%s Read error: %v", remote, err)
		return
	}

	// Echo back line after modification
	response := strings.ToUpper(strings.TrimSpace(message))
	response = fmt.Sprintf("ACK: %s\n", response)
	_, err = conn.Write([]byte(response))
	if err != nil {
		log.Printf("%s Server write error: %v", remote, err)
	}
}
