// In Go, errors are values. That means they are not special and you can program
// them. Here is one programming technique for avoiding repetitive error
// handling. Adapted from https://go.dev/blog/errors-are-values.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type errReader struct {
	r   io.Reader
	err error
}

func (er *errReader) read(buf []byte) {
	// read becomes a no-op as soon as an error occurs
	if er.err != nil {
		return
	}
	_, er.err = er.r.Read(buf)
}

func main() {
	r := strings.NewReader("hello")
	buf := make([]byte, 9)
	er := &errReader{r: r}
	er.read(buf[0:3]) // We do not
	er.read(buf[3:6]) // handle error
	er.read(buf[6:9]) // for each call.
	if er.err != nil {
		fmt.Fprintf(os.Stderr, "err: reading from %#v: %v\n", r, er.err)
		os.Exit(1)
	}
}
