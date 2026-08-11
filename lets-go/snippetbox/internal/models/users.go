package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

// Insert adds a new record to the users table. Email must be unique.
func (m *UserModel) Insert(name, email, password string) error {
	// Create bcrypt hash with cost of 12, ie. 2^12 = 4096 iterations
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("hashing password: %v", err)
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created)
		VALUES(?, ? ,?, UTC_TIMESTAMP())`
	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		// Email not unique?
		if sqlError, ok := errors.AsType[*mysql.MySQLError](err); ok {
			if sqlError.Number == 1062 && strings.Contains(sqlError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}

		return fmt.Errorf("inserting user: %v", err)
	}

	return nil
}

// Authenticate checks user's password and returns ID from users table.
// ErrInvalidCredentials returned if either password is wrong or email not found.
func (m *UserModel) Authenticate(email, password string) (int, error) {
	// Fetch user ID and password hash
	var id int
	var hashedPassword []byte
	stmt := `SELECT id, hashed_password FROM users WHERE email=?`
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, fmt.Errorf("fetching password hash: %v", err)
		}
	}

	// Check password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, fmt.Errorf("checking password: %v", err)
		}
	}

	// Success!
	return id, nil
}

// Exists checks if a user exists with the given ID
func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool
	stmt := "SELECT EXISTS(SELECT true FROM users WHERE id=?)"
	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}
