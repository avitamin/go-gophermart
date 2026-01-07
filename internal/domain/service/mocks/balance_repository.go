// Package mocks provides mock implementations for testing.
package mocks

import (
	"context"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

// BalanceRepository is a mock implementation of repository.BalanceRepository.
type BalanceRepository struct {
	mock.Mock
}

// GetBalance mocks getting balance.
func (m *BalanceRepository) GetBalance(ctx context.Context, userID int64) (*entity.Balance, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Balance), args.Error(1)
}

// CreateWithdrawal mocks creating withdrawal.
func (m *BalanceRepository) CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	args := m.Called(ctx, userID, orderNumber, sum)
	return args.Error(0)
}

// GetWithdrawals mocks getting withdrawals.
func (m *BalanceRepository) GetWithdrawals(ctx context.Context, userID int64) ([]entity.Withdrawal, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Withdrawal), args.Error(1)
}
