package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	readEntireFile()
	openFile()
	seek()
	readAtLeast()
	bufioReader()
}

func openFile() {
	f, err := os.Open("data.txt") // Returns *os.File
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Read into slice of bytes
	buf := make([]byte, 5)
	numRead, err := f.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Read %d bytes: %q\n", numRead, string(buf[:numRead]))
}

func readEntireFile() {
	data, err := os.ReadFile("data.txt") // Returns []byte
	if err != nil {
		panic(err)
	}
	fmt.Print(string(data))
}

func seek() {
	f, err := os.Open("data.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Seek to new offset within file, relative to start of file.
	// Can also seek relative to current offset or file end.
	offset, err := f.Seek(6, io.SeekStart)
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 2)
	numRead, err := f.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Read %d bytes @ %d: %q\n", numRead, offset, string(buf[:numRead]))

	// Rewind to start...
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}

	// ...and to end
	cur, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		panic(err)
	}
	fmt.Printf("End is at %d\n", cur)
}

// readAtLeast from io package is more robust at filling buffers
func readAtLeast() {
	f, err := os.Open("data.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	buf := make([]byte, 2)
	n, err := io.ReadAtLeast(f, buf, 2)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d bytes: %s\n", n, string(buf))
}

// The Reader from the bufio package is more efficient and provides extra methods
func bufioReader() {
	f, err := os.Open("data.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	bs, err := reader.Peek(5)
	fmt.Println(string(bs))
}
