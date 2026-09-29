package main

import (
    "fmt"
    "net/http"
    "time"
)

// Racer returns the URL of the first website to respond.
// An error is returned if a timeout occurs
func Racer(a, b string) (string, error) {
    select {
    case <-ping(a):
        return a, nil
    case <-ping(b):
        return b, nil
    case <-time.After(5 * time.Second):
        return "", fmt.Errorf("timed out waiting for %s and %s", a, b)
    }
}

// ping closes the returned channel after an HTTP GET completes for the given URL
func ping(url string) <-chan struct{} {
    finished := make(chan struct{})
    go func() {
        resp, err := http.Get(url)
        if err == nil {
            resp.Body.Close()
        }
        close(finished)
    }()
    return finished
}
