package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    var wg sync.WaitGroup
    stream := make(chan bool, 1)

    send := func() {
        defer wg.Done()
        fmt.Println("Sender: ready")
        stream <- true
        fmt.Println("Sender: sent")
    }

    receive := func() {
        defer wg.Done()
        fmt.Println("Receiver: not ready yet")
        time.Sleep(100 * time.Millisecond)
        fmt.Println("Receiver: ready")
        <-stream
        fmt.Println("Receiver: received")
    }

    wg.Add(2)
    go send()
    go receive()
    wg.Wait()
}
