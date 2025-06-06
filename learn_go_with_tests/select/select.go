package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	url1 := "https://lost.co.nz/"
	url2 := "https://stuff.co.nz/"
	fmt.Println(Racer(url1, url2))
}

const tenSecondTimeout = 10 * time.Second

func Racer(a, b string) (winner string, error error) {
	return ConfigurableRacer(a, b, tenSecondTimeout)
}

func ConfigurableRacer(a, b string, timeout time.Duration) (winner string, error error) {
	select {
	case <-ping(a):
		return a, nil
	case <-ping(b):
		return b, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("timed out waiting for %q and %q", a, b)
	}
}

func ping(url string) chan struct{} {
	ch := make(chan struct{})
	go func() {
		_, _ = http.Get(url)
		close(ch)
	}()
	return ch
}
