// Package postgres provides PostgreSQL database adapter.
package postgres

import (
	"context"
	"time"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
)

// BalanceRepository implements repository.BalanceRepository for PostgreSQL.
type BalanceRepository struct {
	db *DB
}

// NewBalanceRepository creates a new BalanceRepository.
func NewBalanceRepository(db *DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

// GetBalance retrieves the current balance for a user.
func (r *BalanceRepository) GetBalance(ctx context.Context, userID int64) (*entity.Balance, error) {
	balance := &entity.Balance{}

	// Get total accrued
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(accrual), 0) FROM orders WHERE user_id = $1 AND status = $2",
		userID, entity.OrderStatusProcessed,
	).Scan(&balance.Current)
	if err != nil {
		r.db.LogError("get total accrued", err)
		return nil, err
	}

	// Get total withdrawn
	err = r.db.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1",
		userID,
	).Scan(&balance.Withdrawn)
	if err != nil {
		r.db.LogError("get total withdrawn", err)
		return nil, err
	}

	balance.Current = balance.Current - balance.Withdrawn
	return balance, nil
}

// CreateWithdrawal creates a withdrawal transaction.
func (r *BalanceRepository) CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.db.LogError("begin transaction", err)
		return err
	}
	defer tx.Rollback()

	// Calculate current balance
	var totalAccrued float64
	err = tx.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(accrual), 0) FROM orders WHERE user_id = $1 AND status = $2",
		userID, entity.OrderStatusProcessed,
	).Scan(&totalAccrued)
	if err != nil {
		r.db.LogError("get total accrued in tx", err)
		return err
	}

	var totalWithdrawn float64
	err = tx.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1",
		userID,
	).Scan(&totalWithdrawn)
	if err != nil {
		r.db.LogError("get total withdrawn in tx", err)
		return err
	}

	currentBalance := totalAccrued - totalWithdrawn
	if currentBalance < sum {
		return entity.ErrInsufficientFunds
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO withdrawals (user_id, order_number, sum, processed_at) VALUES ($1, $2, $3, $4)",
		userID, orderNumber, sum, time.Now(),
	)
	if err != nil {
		r.db.LogError("insert withdrawal", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		r.db.LogError("commit transaction", err)
		return err
	}

	return nil
}

// GetWithdrawals retrieves all withdrawals for a user, sorted by time descending.
func (r *BalanceRepository) GetWithdrawals(ctx context.Context, userID int64) ([]entity.Withdrawal, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, order_number, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC`,
		userID,
	)
	if err != nil {
		r.db.LogError("get withdrawals", err)
		return nil, err
	}
	defer rows.Close()

	var withdrawals []entity.Withdrawal
	for rows.Next() {
		var w entity.Withdrawal
		if err := rows.Scan(&w.ID, &w.UserID, &w.OrderNumber, &w.Sum, &w.ProcessedAt); err != nil {
			r.db.LogError("scan withdrawal", err)
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}

	if err := rows.Err(); err != nil {
		r.db.LogError("rows iteration", err)
		return nil, err
	}

	return withdrawals, nil
}
