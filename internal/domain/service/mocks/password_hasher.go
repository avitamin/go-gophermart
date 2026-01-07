// Package mocks provides mock implementations for testing.
package mocks

import (
	"github.com/stretchr/testify/mock"
)

// PasswordHasher is a mock implementation of auth.PasswordHasher.
type PasswordHasher struct {
	mock.Mock
}

// Hash mocks password hashing.
func (m *PasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

// Compare mocks password comparison.
func (m *PasswordHasher) Compare(password, hash string) bool {
	args := m.Called(password, hash)
	return args.Bool(0)
}
