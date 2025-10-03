package main

import (
	"fmt"
	"sync"
)

func main() {
    // Data race! Run with:
    // go run --race temp.go
    var total int
    var wg sync.WaitGroup
    wg.Go(func() { total++ })
    wg.Go(func() { total++ })
    wg.Wait()
    fmt.Println("total:", total)
}
