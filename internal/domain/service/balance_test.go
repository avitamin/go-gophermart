package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/avitamin/go-gophermart/internal/domain/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalanceService_GetBalance(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		balanceRepo := new(mocks.BalanceRepository)
		balance := &entity.Balance{
			Current:   500.5,
			Withdrawn: 42.0,
		}

		balanceRepo.On("GetBalance", ctx, int64(1)).Return(balance, nil)

		service := NewBalanceService(balanceRepo)

		result, err := service.GetBalance(ctx, 1)

		require.NoError(t, err)
		assert.Equal(t, 500.5, result.Current)
		assert.Equal(t, 42.0, result.Withdrawn)
		balanceRepo.AssertExpectations(t)
	})

	t.Run("zero balance", func(t *testing.T) {
		balanceRepo := new(mocks.BalanceRepository)
		balance := &entity.Balance{
			Current:   0,
			Withdrawn: 0,
		}

		balanceRepo.On("GetBalance", ctx, int64(1)).Return(balance, nil)

		service := NewBalanceService(balanceRepo)

		result, err := service.GetBalance(ctx, 1)

		require.NoError(t, err)
		assert.Equal(t, 0.0, result.Current)
		assert.Equal(t, 0.0, result.Withdrawn)
	})

	t.Run("database error", func(t *testing.T) {
		balanceRepo := new(mocks.BalanceRepository)
		dbErr := errors.New("database error")

		balanceRepo.On("GetBalance", ctx, int64(1)).Return(nil, dbErr)

		service := NewBalanceService(balanceRepo)

		_, err := service.GetBalance(ctx, 1)

		assert.Error(t, err)
	})
}

func TestBalanceService_Withdraw(t *testing.T) {
	ctx := context.Background()

	t.Run("successful withdrawal", func(t *testing.T) {
		balanceRepo := new(mocks.BalanceRepository)

		balanceRepo.On("CreateWithdrawal", ctx, int64(1), "79927398713", 100.0).Return(nil)

		service := NewBalanceService(balanceRepo)

		err := service.Withdraw(ctx, 1, &entity.WithdrawRequest{
			Order: "79927398713",
			Sum:   100.0,
		})

		require.NoError(t, err)
		balanceRepo.AssertExpectations(t)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		balanceRepo := new(mocks.BalanceRepository)

		balanceRepo.On("CreateWithdrawal", ctx, int64(1), "79927398713", 1000.0).Return(entity.ErrInsufficientFunds)

		service := NewBalanceService(balanceRepo)

		err := service.Withdraw(ctx, 1, &entity.WithdrawRequest{
			Order: "79927398713",
			Sum:   1000.0,
		})

		assert.ErrorIs(t, err, entity.ErrInsufficientFunds)
	})

	t.Run("invalid order number - empty", func(t *testing.T) {
		balanceRepo := new(mocks.BalanceRepository)

		service := NewBalanceService(balanceRepo)

		err := service.Withdraw(ctx, 1, &entity.WithdrawRequest{
			Order: "",
			Sum:   100.0,
		})

		assert.Error(t, err)
	})

	t.Run("invalid order number - invalid luhn", func(t *testing.T) {
		balanceRepo := new(mocks.BalanceRepository)

		service := NewBalanceService(balanceRepo)

		err := service.Withdraw(ctx, 1, &entity.WithdrawRequest{
			Order: "12345678900", // Invalid Luhn
			Sum:   100.0,
		})

		assert.ErrorIs(t, err, ErrInvalidOrderNumber)
	})

	t.Run("invalid sum - zero", func(t *testing.T) {
		balanceRepo := new(mocks.BalanceRepository)

		service := NewBalanceService(balanceRepo)

		err := service.Withdraw(ctx, 1, &entity.WithdrawRequest{
			Order: "79927398713",
			Sum:   0,
		})

		assert.Error(t, err)
	})

	t.Run("invalid sum - negative", func(t *testing.T) {
		balanceRepo := new(mocks.BalanceRepository)

		service := NewBalanceService(balanceRepo)

		err := service.Withdraw(ctx, 1, &entity.WithdrawRequest{
			Order: "79927398713",
			Sum:   -100.0,
		})

		assert.Error(t, err)
	})
}

func TestBalanceService_GetWithdrawals(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		balanceRepo := new(mocks.BalanceRepository)
		now := time.Now()

		withdrawals := []entity.Withdrawal{
			{
				ID:          1,
				UserID:      1,
				OrderNumber: "2377225624",
				Sum:         500.0,
				ProcessedAt: now,
			},
			{
				ID:          2,
				UserID:      1,
				OrderNumber: "79927398713",
				Sum:         100.0,
				ProcessedAt: now.Add(-time.Hour),
			},
		}

		balanceRepo.On("GetWithdrawals", ctx, int64(1)).Return(withdrawals, nil)

		service := NewBalanceService(balanceRepo)

		result, err := service.GetWithdrawals(ctx, 1)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "2377225624", result[0].OrderNumber)
		balanceRepo.AssertExpectations(t)
	})

	t.Run("empty result", func(t *testing.T) {
		balanceRepo := new(mocks.BalanceRepository)
		balanceRepo.On("GetWithdrawals", ctx, int64(1)).Return([]entity.Withdrawal{}, nil)

		service := NewBalanceService(balanceRepo)

		result, err := service.GetWithdrawals(ctx, 1)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("database error", func(t *testing.T) {
		balanceRepo := new(mocks.BalanceRepository)
		dbErr := errors.New("database error")
		balanceRepo.On("GetWithdrawals", ctx, int64(1)).Return(nil, dbErr)

		service := NewBalanceService(balanceRepo)

		_, err := service.GetWithdrawals(ctx, 1)

		assert.Error(t, err)
	})
}
