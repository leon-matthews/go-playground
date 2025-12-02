package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRacer(t *testing.T) {
	t.Run("compare response time", func(t *testing.T) {
		fastServer := makeDelayedServer(10 * time.Millisecond)
		slowServer := makeDelayedServer(20 * time.Millisecond)
		defer slowServer.Close()
		defer fastServer.Close()

		fastURL := fastServer.URL
		slowURL := slowServer.URL

		want := fastURL
		got, _ := Racer(slowURL, fastURL)

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("timeout after 10s", func(t *testing.T) {
		serverA := makeDelayedServer(20 * time.Millisecond)
		serverB := makeDelayedServer(20 * time.Millisecond)
		defer serverA.Close()
		defer serverB.Close()

		_, err := ConfigurableRacer(serverA.URL, serverB.URL, 10*time.Millisecond)

		if err == nil {
			t.Error("expected an error but did not get one")
		}
	})
}

func makeDelayedServer(delay time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	}))
}
