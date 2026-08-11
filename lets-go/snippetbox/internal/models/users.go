package models

import (
    "database/sql"
    "time"
)

type User struct {
    ID int
    Name string
    Email string
    HashedPassword []byte
    Created time.Time
}

type UserModel struct {
    DB *sql.DB
}

// Insert adds a new record to the users table. Email must be unique.
func (m *UserModel) Insert(name, email, password string) error {
    return nil
}

// Authenticate checks user's password and returns ID from users table.
func (m *UserModel) Authenticate(email, password string) (int, error) {
    return 0, nil
}

// Exists checks if a user exists with the given ID
func (m *UserModel) Exists(id int) (bool, error) {
    return false, nil
}
