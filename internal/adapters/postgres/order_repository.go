// Package postgres provides PostgreSQL database adapter.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/avitamin/go-gophermart/internal/domain/repository"
)

// OrderRepository implements repository.OrderRepository for PostgreSQL.
type OrderRepository struct {
	db *DB
}

// NewOrderRepository creates a new OrderRepository.
func NewOrderRepository(db *DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create creates a new order for a user.
func (r *OrderRepository) Create(ctx context.Context, userID int64, number string) error {
	// Check if order exists
	var existingUserID int64
	err := r.db.QueryRowContext(ctx,
		"SELECT user_id FROM orders WHERE number = $1",
		number,
	).Scan(&existingUserID)

	if err == nil {
		if existingUserID == userID {
			return repository.ErrOrderExists
		}
		return repository.ErrOrderOtherUser
	}

	if !errors.Is(err, sql.ErrNoRows) {
		LogError("check existing order", err)
		return err
	}

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO orders (user_id, number, status, uploaded_at) VALUES ($1, $2, $3, $4)",
		userID, number, entity.OrderStatusNew, time.Now(),
	)
	if err != nil {
		LogError("create order", err)
		return err
	}
	return nil
}

// GetByUserID retrieves all orders for a user, sorted by upload time descending.
func (r *OrderRepository) GetByUserID(ctx context.Context, userID int64) ([]entity.Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, number, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC`,
		userID,
	)
	if err != nil {
		LogError("get orders by user", err)
		return nil, err
	}
	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var order entity.Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			LogError("scan order", err)
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		LogError("rows iteration", err)
		return nil, err
	}

	return orders, nil
}

// GetPending retrieves all orders with non-final status.
func (r *OrderRepository) GetPending(ctx context.Context) ([]entity.Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, number, status, accrual, uploaded_at
		FROM orders
		WHERE status IN ($1, $2)`,
		entity.OrderStatusNew, entity.OrderStatusProcessing,
	)
	if err != nil {
		LogError("get pending orders", err)
		return nil, err
	}
	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var order entity.Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			LogError("scan order", err)
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		LogError("rows iteration", err)
		return nil, err
	}

	return orders, nil
}

// UpdateStatus updates order status and accrual.
func (r *OrderRepository) UpdateStatus(ctx context.Context, number string, status entity.OrderStatus, accrual *float64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE orders SET status = $1, accrual = $2 WHERE number = $3",
		status, accrual, number,
	)
	if err != nil {
		LogError("update order status", err)
		return err
	}
	return nil
}
