package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/avitamin/go-gophermart/internal/domain/repository"
	"github.com/avitamin/go-gophermart/internal/domain/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderService_CreateOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("successful order creation", func(t *testing.T) {
		orderRepo := new(mocks.OrderRepository)
		orderRepo.On("Create", ctx, int64(1), "12345678903").Return(nil)

		service := NewOrderService(orderRepo)

		err := service.CreateOrder(ctx, 1, "12345678903")

		require.NoError(t, err)
		orderRepo.AssertExpectations(t)
	})

	t.Run("order already exists for user", func(t *testing.T) {
		orderRepo := new(mocks.OrderRepository)
		orderRepo.On("Create", ctx, int64(1), "12345678903").Return(repository.ErrOrderExists)

		service := NewOrderService(orderRepo)

		err := service.CreateOrder(ctx, 1, "12345678903")

		assert.ErrorIs(t, err, repository.ErrOrderExists)
		assert.True(t, OrderExistsForUser(err))
	})

	t.Run("order belongs to other user", func(t *testing.T) {
		orderRepo := new(mocks.OrderRepository)
		orderRepo.On("Create", ctx, int64(1), "12345678903").Return(repository.ErrOrderOtherUser)

		service := NewOrderService(orderRepo)

		err := service.CreateOrder(ctx, 1, "12345678903")

		assert.ErrorIs(t, err, repository.ErrOrderOtherUser)
		assert.True(t, OrderBelongsToOther(err))
	})

	t.Run("invalid order number - empty", func(t *testing.T) {
		orderRepo := new(mocks.OrderRepository)

		service := NewOrderService(orderRepo)

		err := service.CreateOrder(ctx, 1, "")

		assert.ErrorIs(t, err, ErrInvalidOrderNumber)
	})

	t.Run("invalid order number - invalid luhn", func(t *testing.T) {
		orderRepo := new(mocks.OrderRepository)

		service := NewOrderService(orderRepo)

		err := service.CreateOrder(ctx, 1, "12345678900") // Invalid Luhn

		assert.ErrorIs(t, err, ErrInvalidOrderNumber)
	})

	t.Run("invalid order number - contains letters", func(t *testing.T) {
		orderRepo := new(mocks.OrderRepository)

		service := NewOrderService(orderRepo)

		err := service.CreateOrder(ctx, 1, "1234abc567")

		assert.ErrorIs(t, err, ErrInvalidOrderNumber)
	})
}

func TestOrderService_GetUserOrders(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		orderRepo := new(mocks.OrderRepository)
		now := time.Now()
		accrual := 100.5

		orders := []entity.Order{
			{
				ID:         1,
				UserID:     1,
				Number:     "12345678903",
				Status:     entity.OrderStatusProcessed,
				Accrual:    &accrual,
				UploadedAt: now,
			},
			{
				ID:         2,
				UserID:     1,
				Number:     "79927398713",
				Status:     entity.OrderStatusNew,
				UploadedAt: now.Add(-time.Hour),
			},
		}

		orderRepo.On("GetByUserID", ctx, int64(1)).Return(orders, nil)

		service := NewOrderService(orderRepo)

		result, err := service.GetUserOrders(ctx, 1)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "12345678903", result[0].Number)
		orderRepo.AssertExpectations(t)
	})

	t.Run("empty result", func(t *testing.T) {
		orderRepo := new(mocks.OrderRepository)
		orderRepo.On("GetByUserID", ctx, int64(1)).Return([]entity.Order{}, nil)

		service := NewOrderService(orderRepo)

		result, err := service.GetUserOrders(ctx, 1)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("database error", func(t *testing.T) {
		orderRepo := new(mocks.OrderRepository)
		dbErr := errors.New("database error")
		orderRepo.On("GetByUserID", ctx, int64(1)).Return(nil, dbErr)

		service := NewOrderService(orderRepo)

		_, err := service.GetUserOrders(ctx, 1)

		assert.Error(t, err)
	})
}

func TestOrderService_GetPendingOrders(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		orderRepo := new(mocks.OrderRepository)
		now := time.Now()

		orders := []entity.Order{
			{
				ID:         1,
				UserID:     1,
				Number:     "12345678903",
				Status:     entity.OrderStatusNew,
				UploadedAt: now,
			},
			{
				ID:         2,
				UserID:     2,
				Number:     "79927398713",
				Status:     entity.OrderStatusProcessing,
				UploadedAt: now,
			},
		}

		orderRepo.On("GetPending", ctx).Return(orders, nil)

		service := NewOrderService(orderRepo)

		result, err := service.GetPendingOrders(ctx)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		orderRepo.AssertExpectations(t)
	})
}

func TestOrderService_UpdateOrderStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		orderRepo := new(mocks.OrderRepository)
		accrual := 500.0

		orderRepo.On("UpdateStatus", ctx, "12345678903", entity.OrderStatusProcessed, &accrual).Return(nil)

		service := NewOrderService(orderRepo)

		err := service.UpdateOrderStatus(ctx, "12345678903", entity.OrderStatusProcessed, &accrual)

		require.NoError(t, err)
		orderRepo.AssertExpectations(t)
	})

	t.Run("update without accrual", func(t *testing.T) {
		orderRepo := new(mocks.OrderRepository)

		orderRepo.On("UpdateStatus", ctx, "12345678903", entity.OrderStatusInvalid, (*float64)(nil)).Return(nil)

		service := NewOrderService(orderRepo)

		err := service.UpdateOrderStatus(ctx, "12345678903", entity.OrderStatusInvalid, nil)

		require.NoError(t, err)
		orderRepo.AssertExpectations(t)
	})
}

func TestOrderExistsForUser(t *testing.T) {
	assert.True(t, OrderExistsForUser(repository.ErrOrderExists))
	assert.False(t, OrderExistsForUser(repository.ErrOrderOtherUser))
	assert.False(t, OrderExistsForUser(errors.New("other error")))
}

func TestOrderBelongsToOther(t *testing.T) {
	assert.True(t, OrderBelongsToOther(repository.ErrOrderOtherUser))
	assert.False(t, OrderBelongsToOther(repository.ErrOrderExists))
	assert.False(t, OrderBelongsToOther(errors.New("other error")))
}
