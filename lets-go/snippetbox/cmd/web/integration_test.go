package main

import (
	"net/http"
	"testing"

	"local.dev/snippetbox/internal/assert"
)

// TestPing is an end-to-end test using out testutils
func TestPing(t *testing.T) {
	app := newApplicationMock(t)
	server := newTestServer(t, app.routes())
	defer server.Close()

	res := server.get(t, "/ping")
	assert.Equal(t, res.status, http.StatusOK)
	assert.Equal(t, res.body, "pong")
}
