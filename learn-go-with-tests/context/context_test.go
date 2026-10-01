package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

type MockStore struct {
	response string
	t        *testing.T
}

func (s *MockStore) Fetch(ctx context.Context) (string, error) {
	// Fetch data (slowly)
	data := make(chan string, 1)
	go func() {
		// One rune at a time, keeping track of ctx.Done()
		var result string
		for _, r := range s.response {
			select {
			case <-ctx.Done():
				log.Println("MockStore was cancelled")
				return
			default:
				time.Sleep(10 * time.Millisecond)
				result += string(r)
			}
		}
		data <- result
	}()

	// Wait for either data or cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case result := <-data:
		return result, nil
	}
}

func TestServer(t *testing.T) {
	t.Run("returns data from the store", func(t *testing.T) {
		want := "Hello, world!"
		store := &MockStore{response: want}
		server := Server(store)
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
		got := response.Body.String()
		assert.Equal(t, want, got)
	})

	t.Run("context times out early", func(t *testing.T) {
		want := "Hello, world!"
		store := &MockStore{response: want}
		server := Server(store)
		response := httptest.NewRecorder()

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx, cancel := context.WithTimeout(request.Context(), 10*time.Millisecond)
		defer cancel()
		request = request.WithContext(ctx)

		server.ServeHTTP(response, request)

		assert.Equal(t, http.StatusGatewayTimeout, response.Code)
		got := response.Body.String()
		t.Log(got)
		assert.Equal(t, "request timeout\n", got)
	})

	t.Run("context cancelled", func(t *testing.T) {
		want := "Hello, world!"
		store := &MockStore{response: want}
		server := Server(store)
		response := httptest.NewRecorder()

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx, cancel := context.WithCancel(request.Context())
		// Call Cancel() before Fetch() finishes
		time.AfterFunc(10*time.Millisecond, cancel)
		request = request.WithContext(ctx)

		server.ServeHTTP(response, request)

		assert.Equal(t, 499, response.Code)
		got := response.Body.String()
		t.Log(got)
		assert.Equal(t, "client cancelled request\n", got)
	})
}
