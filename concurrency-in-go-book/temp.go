package main

import (
	"fmt"
	"sync"
)

func main() {
	salutations := []string{"dad", "mum", "world"}
	var wg sync.WaitGroup

	for _, salutation := range(salutations) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("Hello, %s!\n", salutation)
		}()
	}
	wg.Wait()
}
