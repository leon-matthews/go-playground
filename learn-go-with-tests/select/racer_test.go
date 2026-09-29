package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRacer(t *testing.T) {
	t.Run("return url of fastest server", func(t *testing.T) {
		// synctest.Test(t, func(t *testing.T) {
		t.Log("start servers")
		fastServer := makeServer(t, 10*time.Millisecond)
		slowServer := makeServer(t, 100*time.Millisecond)

		t.Log("run racer")
		want := fastServer.URL
		got, err := Racer(slowServer.URL, fastServer.URL)

		if err != nil {
			t.Errorf("unexpected error")
		}

		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
		// })
	})

	t.Run("return error if no server responds", func(t *testing.T) {
		// synctest.Test(t, func(t *testing.T) {
		fastServer := makeServer(t, 11*time.Second)
		slowServer := makeServer(t, 12*time.Second)

		_, err := Racer(slowServer.URL, fastServer.URL)

		if err == nil {
			t.Errorf("expected an error but didn't get one")
		}
		// })
	})
}

func makeServer(t *testing.T, duration time.Duration) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(duration)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(s.Close)
	return s
}
