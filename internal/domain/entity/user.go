// Package entity contains domain entities for the loyalty system.
package entity

import (
	"errors"
	"time"
)

// ErrPasswordTooShort is returned when password is less than 8 characters.
var ErrPasswordTooShort = errors.New("password must be at least 8 characters")

// User represents a registered user in the loyalty system.
type User struct {
	ID           int64
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}

// UserCredentials represents login credentials for authentication.
type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Validate checks if credentials are valid.
func (c *UserCredentials) Validate() error {
	if c.Login == "" {
		return errors.New("login is required")
	}
	if c.Password == "" {
		return errors.New("password is required")
	}
	if len(c.Password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}
