// Package mocks provides mock implementations for testing.
package mocks

import (
	"context"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

// OrderRepository is a mock implementation of repository.OrderRepository.
type OrderRepository struct {
	mock.Mock
}

// Create mocks order creation.
func (m *OrderRepository) Create(ctx context.Context, userID int64, number string) error {
	args := m.Called(ctx, userID, number)
	return args.Error(0)
}

// GetByUserID mocks getting orders by user ID.
func (m *OrderRepository) GetByUserID(ctx context.Context, userID int64) ([]entity.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Order), args.Error(1)
}

// GetPending mocks getting pending orders.
func (m *OrderRepository) GetPending(ctx context.Context) ([]entity.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Order), args.Error(1)
}

// UpdateStatus mocks updating order status.
func (m *OrderRepository) UpdateStatus(ctx context.Context, number string, status entity.OrderStatus, accrual *float64) error {
	args := m.Called(ctx, number, status, accrual)
	return args.Error(0)
}
