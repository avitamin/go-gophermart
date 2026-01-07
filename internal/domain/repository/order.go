// Package repository defines interfaces for data persistence.
package repository

import (
	"context"
	"errors"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
)

// Order repository errors.
var (
	ErrOrderExists    = errors.New("order already exists for this user")
	ErrOrderOtherUser = errors.New("order belongs to another user")
)

// OrderRepository defines the interface for order data operations.
type OrderRepository interface {
	// Create creates a new order for a user.
	Create(ctx context.Context, userID int64, number string) error

	// GetByUserID retrieves all orders for a user, sorted by upload time descending.
	GetByUserID(ctx context.Context, userID int64) ([]entity.Order, error)

	// GetPending retrieves all orders with non-final status.
	GetPending(ctx context.Context) ([]entity.Order, error)

	// UpdateStatus updates order status and accrual.
	UpdateStatus(ctx context.Context, number string, status entity.OrderStatus, accrual *float64) error
}
