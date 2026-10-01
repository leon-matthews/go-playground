package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// Store defines what a Server needs
type Store interface {
	Fetch(ctx context.Context) (string, error)
}

// Server writes the store's data
func Server(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := store.Fetch(r.Context())
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				http.Error(w, "request timeout", http.StatusGatewayTimeout)
			} else {
				http.Error(w, "client cancelled request", 499)
			}
			return
		}
		fmt.Fprint(w, data)
	}
}
