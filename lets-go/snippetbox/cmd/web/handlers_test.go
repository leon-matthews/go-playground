package main

import (
	"net/http"
	"net/url"
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

	const (
		validName     = "Bob"
		validPassword = "validPa$$word"
		validEmail    = "bob@example.com"
		formTag       = `<form action="/user/signup" method="POST" novalidate>`
	)

	tests := map[string]struct {
		userName          string
		userEmail         string
		userPassword      string
		useValidCSRFToken bool
		wantStatus        int
		wantFormTag       string
	}{
		"Valid submission redirects user": {
			userName:          validName,
			userEmail:         validEmail,
			userPassword:      validPassword,
			useValidCSRFToken: true,
			wantStatus:        http.StatusSeeOther,
		},
		"Invalid CSRF Token causes status bad request": {
			userName:          validName,
			userEmail:         validEmail,
			userPassword:      validPassword,
			useValidCSRFToken: false,
			wantStatus:        http.StatusBadRequest,
		},
		"Empty name causes status unprocessable entity": {
			userName:          "",
			userEmail:         validEmail,
			userPassword:      validPassword,
			useValidCSRFToken: true,
			wantStatus:        http.StatusUnprocessableEntity,
			wantFormTag:       formTag,
		},
		"Empty email causes status unprocessable entity": {
			userName:          validName,
			userEmail:         "",
			userPassword:      validPassword,
			useValidCSRFToken: true,
			wantStatus:        http.StatusUnprocessableEntity,
			wantFormTag:       formTag,
		},
		"Empty password causes status unprocessable entity": {
			userName:          validName,
			userEmail:         validEmail,
			userPassword:      "",
			useValidCSRFToken: true,
			wantStatus:        http.StatusUnprocessableEntity,
			wantFormTag:       formTag,
		},
		"Invalid email causes status unprocessable entity": {
			userName:          validName,
			userEmail:         "I dunno. Hotmail, maybe?",
			userPassword:      validPassword,
			useValidCSRFToken: true,
			wantStatus:        http.StatusUnprocessableEntity,
			wantFormTag:       formTag,
		},
		"Short password causes status unprocessable entity": {
			userName:          validName,
			userEmail:         validEmail,
			userPassword:      "1234",
			useValidCSRFToken: true,
			wantStatus:        http.StatusUnprocessableEntity,
			wantFormTag:       formTag,
		},
		"Duplicate email causes status unprocessable entity": {
			userName:          validName,
			userEmail:         "dupe@example.com",
			userPassword:      validPassword,
			useValidCSRFToken: true,
			wantStatus:        http.StatusUnprocessableEntity,
			wantFormTag:       formTag,
		},
	}

	const urlPath = "/user/signup"
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Reset the cookie jar for each sub-test.
			server.clearCookies(t)

			// GET form
			// Adds CSRF cookie to client, and response contains CSRF token
			res := server.get(t, urlPath)

			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			if tt.useValidCSRFToken {
				form.Add("csrf_token", extractCSRFToken(t, res.body))
			}

			// POST form
			response := server.postForm(t, urlPath, form)

			// And finally, test the response data.
			assert.Equal(t, response.status, tt.wantStatus)
			assert.True(t, strings.Contains(response.body, tt.wantFormTag))
		})
	}

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
