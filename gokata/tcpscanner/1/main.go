// Tcpscanner reports open TCP ports on a host. First create a pool of workers
// that will do the scanning by connecting to ports. Then send them port numbers
// to try to connect to. Collect the results, 0 means couldn't connect, and
// print them. Adapted from the "Black Hat Go" [book].
//
// Topics: concurrency, security
// Level: intermediate
//
// [book]: https://github.com/blackhat-go/bhg/blob/master/ch-2/tcp-scanner-final
package main

import (
	"fmt"
	"net"
	"sort"
)

const (
	host           = "scanme.nmap.org"
	portRangeStart = 1
	portRangeEnd   = 1024
	nWorkers       = 100
)

func worker(portsToScan, portsScanned chan int) {
	for port := range portsToScan {
		addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			// closed port		syn->, <-rst
			// filtered port	syn->, timeout
			portsScanned <- 0
			continue
		}
		conn.Close()
		portsScanned <- port
	}
}

func main() {
	portsToScan := make(chan int, nWorkers) // can hold n items before sender blocks
	portsScanned := make(chan int)

	for range nWorkers {
		go worker(portsToScan, portsScanned)
	}

	go func() {
		for i := portRangeStart; i <= portRangeEnd; i++ {
			portsToScan <- i
		}
	}()

	var openPorts []int

	for i := portRangeStart; i <= portRangeEnd; i++ {
		port := <-portsScanned
		if port != 0 {
			openPorts = append(openPorts, port)
		}
		fmt.Printf("\r%d", i) // show progress
	}
	fmt.Printf("\n")

	sort.Ints(openPorts)
	fmt.Println(host, openPorts)
}
