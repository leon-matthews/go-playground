package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"local.dev/snippetbox/internal/models/mocks"
)

// newApplicationMock creates instance of our app with mocked dependencies
func newApplicationMock(t *testing.T) *application {
	// Template cache
	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	// Session manager
	// No store specified, so SCS defaults to in-memory.
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	// Mock application, with mock models.
	return &application{
		formDecoder:    form.NewDecoder(),
		logger:         slog.New(slog.DiscardHandler),
		sessionManager: sessionManager,
		snippets:       &mocks.SnippetModel{},
		templates:      templateCache,
		users:          &mocks.UserModel{},
	}
}

// testResponse holds data about responses from the test server
type testResponse struct {
	status  int
	headers http.Header
	cookies []*http.Cookie
	body    string
}

// testServer is a custom type that embeds an httptest.Server instance
type testServer struct {
	*httptest.Server
}

// newTestServer creates a TLS server using the given handler
func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	// Add cookie jar to server's client
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	ts.Client().Jar = jar

	// Prevent automatic following of redirects by server's client
	ts.Client().CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

// get makes a GET request using the test server client
func (ts *testServer) get(t *testing.T, urlPath string) testResponse {
	request, err := http.NewRequest(http.MethodGet, ts.URL+urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	response, err := ts.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	return testResponse{
		status:  response.StatusCode,
		headers: response.Header,
		cookies: response.Cookies(),
		body:    string(bytes.TrimSpace(body)),
	}
}

// clearCookies reset test server's client to a new and empty cookie jar.
func (ts *testServer) clearCookies(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	ts.Client().Jar = jar
}
