package main

import (
	"fmt"
	"os"
)

func main() {
	f := createFile("/tmp/defer.txt")
	defer closeFile(f)
	writeFile(f)
}

func createFile(path string) *os.File {
	fmt.Printf("Creating: %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	return f
}

func closeFile(f *os.File) {
    fmt.Println("closing")
    err := f.Close()
    if err != nil {
        panic(err)
    }
}

func writeFile(f *os.File) {
	fmt.Println("writing")
    fmt.Fprintln(f, "data")
}
