package models

import (
	"errors"
)

var (
	// Snippets
	ErrNoRecord = errors.New("models: no matching record found")

	// Users
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail     = errors.New("models: duplicate email")
)
