// Command tcpscanner concurrently probes a range of TCP ports on a host.
package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

func main() {
	host := "scanme.nmap.org"
	openports := scan(host, 1, 32_000, 100, 500*time.Millisecond)
	for port := range openports {
		fmt.Println(port)
	}
}

// scanner carries the shared state used by every port-probing goroutine.
type scanner struct {
	host      string
	timeout   time.Duration
	openports chan int
	semaphore chan struct{}
	wg        sync.WaitGroup
}

// scan probes ports portStart..portEnd on host, up to concurrency at once.
//
// It returns a channel that yields each open port as it is found, closed
// once every port has been probed.
func scan(host string, portStart, portEnd, concurrency int, timeout time.Duration) <-chan int {
	s := &scanner{
		host:      host,
		timeout:   timeout,
		openports: make(chan int),
		semaphore: make(chan struct{}, concurrency),
	}
	go s.run(portStart, portEnd)
	return s.openports
}

// run spawns one worker per port and closes openports once all have finished.
func (s *scanner) run(portStart, portEnd int) {
	for port := portStart; port <= portEnd; port++ {
		s.semaphore <- struct{}{} // blocks once concurrency workers are busy
		s.wg.Add(1)
		go s.scanPort(port)
	}
	s.wg.Wait()
	close(s.openports)
}

// scanPort probes a single port and reports it on openports when open.
func (s *scanner) scanPort(port int) {
	defer s.wg.Done()
	defer func() { <-s.semaphore }() // free the slot for the next port
	if connected(s.host, port, s.timeout) {
		s.openports <- port
	}
}

// connected reports whether a TCP connection to host:port succeeds within timeout.
func connected(host string, port int, timeout time.Duration) bool {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err == nil {
		conn.Close()
		return true
	}
	return false
}
