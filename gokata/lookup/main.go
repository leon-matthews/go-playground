// Lookup looks up IP addresses concurrently. Hosts are supplied as STDIN stream.
// Adapted from https://youtu.be/woCg2zaIVzQ.
package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

const numWorkers = 10

type lookup struct {
	host  string
	addrs []string
	err   error
}

func main() {
	hosts := readHosts(os.Stdin)
	results := lookupHosts(hosts)

	// Print errors on stderr and lookups on stdout.
	for lookup := range results {
		if lookup.err != nil {
			fmt.Fprintf(os.Stderr, "lookup: %v\n", lookup.err)
		} else {
			fmt.Printf("%s: %s\n", lookup.host, strings.Join(lookup.addrs, ", "))
		}
	}
}

// readHosts pushes lines from stdin into the returned channel
func readHosts(r io.Reader) <-chan string {
	hosts := make(chan string)
	go func() {
		defer close(hosts)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			hosts <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "reading: %v\n", err)
			os.Exit(1)
		}
	}()
	return hosts
}

func lookupHosts(hosts <-chan string) <-chan lookup {
	// Spin-up workers to look-up hosts
	lookups := make(chan lookup)

	// Each worker pulls work from hosts (until its closed) then writes results to lookups.
	var wg sync.WaitGroup
	for range numWorkers {
		wg.Go(func() {
			for host := range hosts {
				addrs, err := net.LookupHost(host)
				lookups <- lookup{host: host, addrs: addrs, err: err}
			}
		})
	}

	// Close lookups once all the works are finished
	go func() {
		wg.Wait()
		close(lookups)
	}()

	return lookups
}
