package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrderStatus_IsFinal(t *testing.T) {
	tests := []struct {
		name   string
		status OrderStatus
		want   bool
	}{
		{"NEW is not final", OrderStatusNew, false},
		{"PROCESSING is not final", OrderStatusProcessing, false},
		{"INVALID is final", OrderStatusInvalid, true},
		{"PROCESSED is final", OrderStatusProcessed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsFinal())
		})
	}
}

func TestOrder_ToResponse(t *testing.T) {
	now := time.Now()
	accrual := 100.5

	order := Order{
		ID:         1,
		UserID:     2,
		Number:     "12345678903",
		Status:     OrderStatusProcessed,
		Accrual:    &accrual,
		UploadedAt: now,
	}

	response := order.ToResponse()

	assert.Equal(t, order.Number, response.Number)
	assert.Equal(t, order.Status, response.Status)
	assert.Equal(t, order.Accrual, response.Accrual)
	assert.Equal(t, order.UploadedAt, response.UploadedAt)
}

func TestOrder_ToResponseWithoutAccrual(t *testing.T) {
	now := time.Now()

	order := Order{
		ID:         1,
		UserID:     2,
		Number:     "12345678903",
		Status:     OrderStatusNew,
		Accrual:    nil,
		UploadedAt: now,
	}

	response := order.ToResponse()

	assert.Equal(t, order.Number, response.Number)
	assert.Equal(t, order.Status, response.Status)
	assert.Nil(t, response.Accrual)
}
