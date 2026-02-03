package main

import (
	"log"
	"time"
)

func main() {
	leakyTickerExample()
	goroutinelessAfterFuncExample()
}

// afterFuncExample() ----------------------------------------------------------
func goroutinelessAfterFuncExample() {
	log.Println("Starting goroutine-less example using time.AfterFunc()")
	t := newAfterFuncExample(time.Second * 1)
	t.loop()

	time.Sleep(3 * time.Second)
	log.Println("Send stop")
	t.stop()
}

type afterFuncExample struct {
	latency time.Duration
	t       *time.Timer
}

func newAfterFuncExample(latency time.Duration) *afterFuncExample {
	return &afterFuncExample{
		latency: latency,
		t:       nil,
	}
}

func (t *afterFuncExample) action() {
	log.Println("Action ticked")
}

func (t *afterFuncExample) loop() {
	if t.t == nil {
		t.t = time.AfterFunc(t.latency, t.action)
	} else {
		t.t.Reset(t.latency)
	}
}

func (t *afterFuncExample) stop() {
	t.t.Stop()
}

// tickerExample ---------------------------------------------------------------
func leakyTickerExample() {
	log.Println("Starting leaky ticker")
	t := newTickerExample(time.Second * 1)
	go t.loop()

	// If `t.stop()` is never called, the goroutine above will never exit
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
