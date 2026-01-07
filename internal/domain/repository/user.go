// Package repository defines interfaces for data persistence.
package repository

import (
	"context"
	"errors"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
)

// User repository errors.
var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	// Create creates a new user and returns their ID.
	Create(ctx context.Context, login, passwordHash string) (int64, error)

	// GetByLogin retrieves a user by their login.
	GetByLogin(ctx context.Context, login string) (*entity.User, error)

	// GetByID retrieves a user by their ID.
	GetByID(ctx context.Context, id int64) (*entity.User, error)
}
