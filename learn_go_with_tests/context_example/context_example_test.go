package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func NewSpyStore(data string) *SpyStore {
	return &SpyStore{
		cancelled: false,
		response:  data,
	}
}

type SpyStore struct {
	cancelled bool
	response  string
}

func (s *SpyStore) Cancel() {
	s.cancelled = true
}

func (s *SpyStore) Fetch() string {
	time.Sleep(100 * time.Millisecond)
	return s.response
}

func TestServer(t *testing.T) {
	t.Run("easy case", func(t *testing.T) {
		data := "Hello, world!"
		store := NewSpyStore(data)
		server := Server(store)

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		if response.Body.String() != data {
			t.Errorf(`got "%s", want "%s"`, response.Body.String(), data)
		}

		if store.cancelled {
			t.Error("store should not be cancelled")
		}
	})

	t.Run("timeout", func(t *testing.T) {
		data := "Hello, world!"
		store := NewSpyStore(data)
		server := Server(store)

		request := httptest.NewRequest(http.MethodGet, "/", nil)

		cancellingCtx, cancel := context.WithCancel(request.Context())
		time.AfterFunc(5*time.Millisecond, cancel)
		request = request.WithContext(cancellingCtx)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		if !store.cancelled {
			t.Error("store should be cancelled")
		}

		if response.Body.String() != data {
			t.Errorf(`got "%s", want "%s"`, response.Body.String(), data)
		}
	})
}
