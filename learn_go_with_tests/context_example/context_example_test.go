package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type SpyStore struct {
	cancelled bool
	response  string
	t         *testing.T
}

func NewSpyStore(t *testing.T, data string) *SpyStore {
	return &SpyStore{
		cancelled: false,
		response:  data,
		t:         t,
	}
}

func (s *SpyStore) assertCancelled() {
	s.t.Helper()
	if !s.cancelled {
		s.t.Error("store was not told to cancel")
	}
}

func (s *SpyStore) assertNotCancelled() {
	s.t.Helper()
	if s.cancelled {
		s.t.Error("store was told to cancel")
	}
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
		store := NewSpyStore(t, data)
		server := Server(store)

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		store.assertNotCancelled()
		if response.Body.String() != data {
			t.Errorf(`got "%s", want "%s"`, response.Body.String(), data)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		data := "Hello, world!"
		store := NewSpyStore(t, data)
		server := Server(store)

		request := httptest.NewRequest(http.MethodGet, "/", nil)

		cancellingCtx, cancel := context.WithCancel(request.Context())
		time.AfterFunc(5*time.Millisecond, cancel)
		request = request.WithContext(cancellingCtx)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		store.assertCancelled()
	})
}
