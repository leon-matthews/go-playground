package main

import (
	"log"
	"time"
)

func main() {
	leakyTickerExample()
}

func leakyTickerExample() {
	t := newTickerExample(time.Second * 1)
	go t.loop()

	// If stop is never called, the goroutine above will never exit
	time.Sleep(3 * time.Second)
	log.Println("Send stop")
	t.stop()
}

type tickerExample struct {
	latency time.Duration
	done    chan bool
}

func newTickerExample(latency time.Duration) *tickerExample {
	return &tickerExample{
		latency: latency,
		done:    make(chan bool),
	}
}

func (t *tickerExample) loop() {
	ticker := time.NewTicker(t.latency)
	defer ticker.Stop()
	for {
		select {
		case <-t.done:
			return
		case <-ticker.C:
			log.Println("Ticker example ticked")
		}
	}
}

func (t *tickerExample) stop() {
	t.done <- true
}
