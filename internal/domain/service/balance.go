// Package service contains business logic for the loyalty system.
package service

import (
	"context"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/avitamin/go-gophermart/internal/domain/repository"
	"github.com/avitamin/go-gophermart/internal/pkg/luhn"
)

// BalanceService handles balance-related business logic.
type BalanceService struct {
	repo repository.BalanceRepository
}

// NewBalanceService creates a new BalanceService.
func NewBalanceService(repo repository.BalanceRepository) *BalanceService {
	return &BalanceService{repo: repo}
}

// GetBalance retrieves the current balance for a user.
func (s *BalanceService) GetBalance(ctx context.Context, userID int64) (*entity.Balance, error) {
	return s.repo.GetBalance(ctx, userID)
}

// Withdraw creates a withdrawal transaction.
func (s *BalanceService) Withdraw(ctx context.Context, userID int64, req *entity.WithdrawRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	if !luhn.Valid(req.Order) {
		return ErrInvalidOrderNumber
	}

	return s.repo.CreateWithdrawal(ctx, userID, req.Order, req.Sum)
}

// GetWithdrawals retrieves all withdrawals for a user.
func (s *BalanceService) GetWithdrawals(ctx context.Context, userID int64) ([]entity.Withdrawal, error) {
	return s.repo.GetWithdrawals(ctx, userID)
}
