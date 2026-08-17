package main

import (
	"net/http"
	"strings"
	"testing"

	"local.dev/snippetbox/internal/assert"
)

// TestPing is an end-to-end test using out testutils
func TestPing(t *testing.T) {
	app := newApplicationMock(t)
	server := newTestServer(t, app.routes())
	defer server.Close()

	response := server.get(t, "/ping")
	assert.Equal(t, response.status, http.StatusOK)
	assert.Equal(t, response.body, "pong")
}

func TestUserSignup(t *testing.T) {
	app := newApplicationMock(t)
	server := newTestServer(t, app.routes())
	defer server.Close()

	// Fetch form to get CSRF token
	response := server.get(t, "/user/signup")
	assert.Equal(t, response.status, http.StatusOK)
	token := extractCSRFToken(t, response.body)
	t.Logf("CSRF token: %s", token)
	t.Logf("Cookies: %s", response.cookies)
}

func TestSnippetView(t *testing.T) {
	t.Parallel()
	app := newApplicationMock(t)
	server := newTestServer(t, app.routes())
	defer server.Close()

	tests := map[string]struct {
		urlPath    string
		wantStatus int
		wantBody   string
	}{
		"Valid ID": {
			urlPath:    "/snippet/view/1",
			wantStatus: http.StatusOK,
			wantBody:   "An old silent pond...",
		},
		"Empty ID": {
			urlPath:    "/snippet/view/",
			wantStatus: http.StatusNotFound,
		},
		"Invalid ID": {
			urlPath:    "/snippet/view/2",
			wantStatus: http.StatusNotFound,
		},
		"Negative ID": {
			urlPath:    "/snippet/view/-1",
			wantStatus: http.StatusNotFound,
		},
		"Decimal number": {
			urlPath:    "/snippet/view/1.23",
			wantStatus: http.StatusNotFound,
		},
		"String ID": {
			urlPath:    "/snippet/view/banana",
			wantStatus: http.StatusNotFound,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// server is shared, subtests not parallelisable
			server.clearCookies(t)
			response := server.get(t, tt.urlPath)
			assert.Equal(t, response.status, tt.wantStatus)
			assert.True(t, strings.Contains(response.body, tt.wantBody))
		})
	}
}
