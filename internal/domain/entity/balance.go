// Package entity contains domain entities for the loyalty system.
package entity

import (
	"errors"
	"time"
)

// Balance errors.
var (
	ErrInsufficientFunds = errors.New("insufficient funds")
)

// Balance represents user's loyalty points balance.
type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// CanWithdraw checks if the balance has enough funds for withdrawal.
func (b *Balance) CanWithdraw(amount float64) bool {
	return b.Current >= amount
}

// Withdrawal represents a withdrawal transaction.
type Withdrawal struct {
	ID          int64
	UserID      int64
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}

// WithdrawalResponse represents withdrawal data for API response.
type WithdrawalResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

// ToResponse converts Withdrawal to WithdrawalResponse.
func (w *Withdrawal) ToResponse() WithdrawalResponse {
	return WithdrawalResponse{
		Order:       w.OrderNumber,
		Sum:         w.Sum,
		ProcessedAt: w.ProcessedAt,
	}
}

// WithdrawRequest represents a withdrawal request from API.
type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// Validate checks if the withdrawal request is valid.
func (r *WithdrawRequest) Validate() error {
	if r.Order == "" {
		return errors.New("order number is required")
	}
	if r.Sum <= 0 {
		return errors.New("sum must be positive")
	}
	return nil
}
