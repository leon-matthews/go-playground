// Ping ponging messages between two goroutines on a channel, measuring number
// of messages sent per second. This should be roughly equivalent to the
// thread-pipe-msgpersec.c benchmark.
//
// Eli Bendersky [http://eli.thegreenplace.net]
// This code is in the public domain.
package main

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
)

// child takes a channel c and loops over messages in it until it's closed,
// echoing all the messages back into the channel.
func child(c chan string) {
	for msg := range c {
		if len(msg) != 4 {
			panic("unexpected message length")
		}
		c <- msg
	}
}

func main() {
	c := make(chan string)
	go child(c)

	t1 := time.Now()
	const niters = 20_000_000
	for i := 0; i < niters; i++ {
		c <- "leon"
		reply := <-c
		if "leon" != reply {
			panic("oh no, mismatch")
		}
	}
	elapsed := time.Since(t1)

	per_sec := int64(niters / elapsed.Seconds())
	fmt.Printf(
		"%s messages total, averaging %v each. %s iters/sec\n",
		humanize.Comma(niters),
		elapsed/niters,
		humanize.Comma(per_sec),
	)
}
