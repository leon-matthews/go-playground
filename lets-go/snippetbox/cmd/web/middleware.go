package main

import (
	"net/http"
	"runtime"
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
