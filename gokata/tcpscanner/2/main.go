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
	openports := scan(host, 1, 1024, 100, 500*time.Millisecond)
	for port := range openports {
		fmt.Println(port)
	}
}

func scan(host string, portStart, portEnd, concurrency int, timeout time.Duration) <-chan int {
	openports := make(chan int)
	go func() {
		semaphore := make(chan struct{}, concurrency)
		var wg sync.WaitGroup
		for i := portStart; i <= portEnd; i++ {
			semaphore <- struct{}{}
			wg.Add(1)
			go func(port int) {
				if connected(host, port, timeout) {
					openports <- port
				}
				<-semaphore
				wg.Done()
			}(i)
		}
		wg.Wait()
		close(openports)
		close(semaphore)
	}()
	return openports
}

func connected(host string, port int, timeout time.Duration) bool {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err == nil {
		conn.Close()
		return true
	}
	return false
}
