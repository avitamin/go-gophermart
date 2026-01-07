// Package mocks provides mock implementations for testing.
package mocks

import (
	"context"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

// UserRepository is a mock implementation of repository.UserRepository.
type UserRepository struct {
	mock.Mock
}

// Create mocks user creation.
func (m *UserRepository) Create(ctx context.Context, login, passwordHash string) (int64, error) {
	args := m.Called(ctx, login, passwordHash)
	return args.Get(0).(int64), args.Error(1)
}

// GetByLogin mocks getting user by login.
func (m *UserRepository) GetByLogin(ctx context.Context, login string) (*entity.User, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

// GetByID mocks getting user by ID.
func (m *UserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
