package mocks

import "local.dev/snippetbox/internal/models"

// UserModel mocks database access to users table
type UserModel struct{}

// Insert mock implemenation.
// Use "dupe@example.com" to get duplicate email error.
func (m *UserModel) Insert(name, email, password string) error {
	switch email {
	case "dupe@example.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

// Authenticate mock implemenation.
// For a valid login use: alice@example.com/pa$$word
func (m *UserModel) Authenticate(email, password string) (int, error) {
	if email == "alice@example.com" && password == "pa$$word" {
		return 1, nil
	}

	return 0, models.ErrInvalidCredentials
}

// Exists mock implementation
// Use ID of 1 for true response
func (m *UserModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}
