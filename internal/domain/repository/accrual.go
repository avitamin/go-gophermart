// Package repository defines interfaces for data persistence.
package repository

import (
	"context"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
)

// AccrualResponse represents the response from accrual system.
type AccrualResponse struct {
	Order   string
	Status  entity.OrderStatus
	Accrual *float64
}

// AccrualClient defines the interface for accrual system communication.
type AccrualClient interface {
	// GetOrderAccrual retrieves accrual information for an order.
	// Returns nil response if order is not registered in accrual system.
	// Returns error with retry duration for rate limiting (429).
	GetOrderAccrual(ctx context.Context, orderNumber string) (*AccrualResponse, error)
}
