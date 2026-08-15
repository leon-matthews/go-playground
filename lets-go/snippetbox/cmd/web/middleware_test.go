package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"local.dev/snippetbox/internal/assert"
)

func TestCommonHeaders(t *testing.T) {
	// Create ResponseRecorder and a dummy request
	rr := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create mock HTTP Handler
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("OK"))
	})

	// Run middleware
	commonHeaders(next).ServeHTTP(rr, req)

	// Check response...
	response := rr.Result()
	defer response.Body.Close()

	// ...status code
	assert.Equal(t, response.StatusCode, http.StatusOK)

	// ...body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)
	assert.Equal(t, string(body), "OK")

	// ...headers
	var want string
	want = "default-src 'self'; style-src 'self'; font-src 'self'"
	assert.Equal(t, response.Header.Get("Content-Security-Policy"), want)

	// Check that the middleware has correctly set the Referrer-Policy
	// header on the response.
	want = "origin-when-cross-origin"
	assert.Equal(t, response.Header.Get("Referrer-Policy"), want)

	// Check that the middleware has correctly set the X-Content-Type-Options
	// header on the response.
	want = "nosniff"
	assert.Equal(t, response.Header.Get("X-Content-Type-Options"), want)

	// Check that the middleware has correctly set the X-Frame-Options header
	// on the response.
	want = "deny"
	assert.Equal(t, response.Header.Get("X-Frame-Options"), want)

	// Check that the middleware has correctly set the X-XSS-Protection header
	// on the response
	want = "0"
	assert.Equal(t, response.Header.Get("X-XSS-Protection"), want)

	// Check that the middleware has correctly set the Server header on the
	// response.
	want = "Go 1."
	assert.Equal(t, response.Header.Get("Server")[:5], want)
}
