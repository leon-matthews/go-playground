package main

import (
	"bytes"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
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
	// Request
	request, err := http.NewRequest(http.MethodGet, ts.URL+urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	return ts.clientDo(t, request)
}

// postForm sends a default 'application/x-www-form-urlencode' form to given urlPath
func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) testResponse {
	// Request
	request, err := http.NewRequest(http.MethodPost, ts.URL+urlPath, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	return ts.clientDo(t, request)
}

// clientDo has the test server handle the request and builds a response
func (ts *testServer) clientDo(t *testing.T, request *http.Request) testResponse {
	// Response
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

// extractCSRFToken extracts the value from the hidden CSRF form field from given HTML
func extractCSRFToken(t *testing.T, body string) string {
	// Regex which captures the CSRF token value from an HTML form
	csrfTokenRX := regexp.MustCompile(`<input type="hidden" name="csrf_token" value="(.+)">`)
	matches := csrfTokenRX.FindStringSubmatch(body)
	if len(matches) < 2 {
		t.Fatal("no csrf token found in body")
	}
	// Unescape match in case taken contains a plus: '&#43;' unescaped to '+'
	return html.UnescapeString(matches[1])
}
