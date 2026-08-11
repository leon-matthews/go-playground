package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime"

	"github.com/justinas/nosurf"
)

var goVersionString = "Go " + runtime.Version()[2:]

func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Server", goVersionString)

		// Content Security Policy (CSP) header
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/CSP
		h.Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; font-src 'self'")

		// Strip path & query from referrer, if not on same origin
		h.Set("Referrer-Policy", "origin-when-cross-origin")

		// Help prevent content-sniffing attacks by instructing browsers to not guess MIME-type
		h.Set("X-Content-Type-Options", "nosniff")

		// Help prevent clickjacking attacks in older browsers that do not support CSP headers
		h.Set("X-Frame-Options", "deny")

		// Disable browser's built-in XSS filtering, as we are using CSP headers
		h.Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

// preventCSRF uses a CSRF token check
func preventCSRF(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})
	return csrfHandler
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.logger.Debug("received request",
			slog.String("ip", r.RemoteAddr),
			slog.String("proto", r.Proto),
			slog.String("method", r.Method),
			slog.String("uri", r.URL.RequestURI()),
		)
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// Use the built-in recover() function to check if a panic occurred.
			pv := recover()

			// If a panic did happen...
			if pv != nil {
				// ...set a "Connection: close" header on the response, then send 500 error
				w.Header().Set("Connection", "close")
				err := fmt.Errorf("Panic in handler: %v", pv)
				app.serverError(w, r, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Redirect, then return to prevent subsequent handlers from running
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		// Don't cache protected pages
		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
