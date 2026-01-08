// Package repository defines interfaces for data persistence.
package repository

import (
	"context"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
)

// BalanceRepository defines the interface for balance and withdrawal operations.
type BalanceRepository interface {
	// GetBalance retrieves the current balance for a user.
	GetBalance(ctx context.Context, userID int64) (*entity.Balance, error)

	// CreateWithdrawal creates a withdrawal transaction.
	// Returns entity.ErrInsufficientFunds if balance is insufficient.
	CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error

	// GetWithdrawals retrieves all withdrawals for a user, sorted by time descending.
	GetWithdrawals(ctx context.Context, userID int64) ([]entity.Withdrawal, error)
}
