package models

import (
	"errors"
)

var (
	// ErrNoRecord indicates a failed fetch
	ErrNoRecord = errors.New("models: no matching record found")

	// ErrInvalidCredentials indicates either password or email lookup failed
	ErrInvalidCredentials = errors.New("models: invalid credentials")

	// ErrDuplicateEmail indicates user record already exists
	ErrDuplicateEmail = errors.New("models: duplicate email")
)
