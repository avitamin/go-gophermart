// Package entity contains domain entities for the loyalty system.
package entity

import (
	"time"
)

// OrderStatus represents the processing status of an order.
type OrderStatus string

// Order status constants.
const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

// IsFinal returns true if the status is final and won't change.
func (s OrderStatus) IsFinal() bool {
	return s == OrderStatusInvalid || s == OrderStatusProcessed
}

// Order represents a user's order in the loyalty system.
type Order struct {
	ID         int64
	UserID     int64
	Number     string
	Status     OrderStatus
	Accrual    *float64
	UploadedAt time.Time
}

// OrderResponse represents order data for API response.
type OrderResponse struct {
	Number     string      `json:"number"`
	Status     OrderStatus `json:"status"`
	Accrual    *float64    `json:"accrual,omitempty"`
	UploadedAt time.Time   `json:"uploaded_at"`
}

// ToResponse converts Order to OrderResponse.
func (o *Order) ToResponse() OrderResponse {
	return OrderResponse{
		Number:     o.Number,
		Status:     o.Status,
		Accrual:    o.Accrual,
		UploadedAt: o.UploadedAt,
	}
}
