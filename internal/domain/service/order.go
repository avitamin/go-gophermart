// Package service contains business logic for the loyalty system.
package service

import (
	"context"
	"errors"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/avitamin/go-gophermart/internal/domain/repository"
	"github.com/avitamin/go-gophermart/internal/pkg/luhn"
)

// OrderService errors.
var (
	ErrInvalidOrderNumber = errors.New("invalid order number")
)

// OrderService handles order-related business logic.
type OrderService struct {
	repo repository.OrderRepository
}

// NewOrderService creates a new OrderService.
func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// CreateOrder creates a new order for a user.
// Returns repository.ErrOrderExists if order was already submitted by this user.
// Returns repository.ErrOrderOtherUser if order belongs to another user.
func (s *OrderService) CreateOrder(ctx context.Context, userID int64, orderNumber string) error {
	if orderNumber == "" || !luhn.Valid(orderNumber) {
		return ErrInvalidOrderNumber
	}

	return s.repo.Create(ctx, userID, orderNumber)
}

// GetUserOrders retrieves all orders for a user.
func (s *OrderService) GetUserOrders(ctx context.Context, userID int64) ([]entity.Order, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// GetPendingOrders retrieves all orders awaiting processing.
func (s *OrderService) GetPendingOrders(ctx context.Context) ([]entity.Order, error) {
	return s.repo.GetPending(ctx)
}

// UpdateOrderStatus updates order status and accrual.
func (s *OrderService) UpdateOrderStatus(ctx context.Context, number string, status entity.OrderStatus, accrual *float64) error {
	return s.repo.UpdateStatus(ctx, number, status, accrual)
}

// OrderExistsForUser checks if the error indicates the order exists for current user.
func OrderExistsForUser(err error) bool {
	return errors.Is(err, repository.ErrOrderExists)
}

// OrderBelongsToOther checks if the error indicates the order belongs to another user.
func OrderBelongsToOther(err error) bool {
	return errors.Is(err, repository.ErrOrderOtherUser)
}
