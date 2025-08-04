package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	writeFile()
	create()
}

// For more granular control use [os.Create]
func create() {
	f, err := os.CreateTemp("", "writing-files-*")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	// Write bytes
	data := []byte{115, 111, 109, 101, 10}
	num, err := f.Write(data)
	if err != nil {
		panic(err)
	}
	fmt.Printf("wrote %d bytes\n", num)

	// Write string
	num, err = f.WriteString("writes\n")
	if err != nil {
		panic(err)
	}
	fmt.Printf("wrote %d bytes\n", num)

	// Write to storage
	f.Sync()

	// bufio also provides a writer
	writer := bufio.NewWriter(f)
	num, err = writer.WriteString("buffered")
	if err != nil {
		panic(err)
	}
	fmt.Printf("wrote %d bytes\n", num)

	// Write to storage
	writer.Flush()

	// Read it back
	bs, err := os.ReadFile(f.Name())
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bs))
}

// WriteFile writes data to the named file in one step, creating and truncating it if necessary.
func writeFile() {
	data := []byte("Hello, Go!\n")
	err := os.WriteFile("data2.txt", data, 0666)
	if err != nil {
		panic(err)
	}
	fmt.Printf("wrote %d bytes\n", len(data))
}
