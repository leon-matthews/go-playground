package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"local.dev/snippetbox/internal/assert"
)

// TestPing is an end-to-end smoke test
func TestPing(t *testing.T) {
	// Minimal instance of application struct
	app := &application{
		logger: slog.New(slog.DiscardHandler),
	}

	// Create test server
	server := httptest.NewTLSServer(app.routes())
	defer server.Close()

	// Create request
	url := server.URL + "/ping"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Execute request against the test server
	client := server.Client()
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	// Check status
	assert.Equal(t, response.StatusCode, http.StatusOK)

	// Check body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)
	assert.Equal(t, string(body), "pong")
}
