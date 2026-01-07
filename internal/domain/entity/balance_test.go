package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBalance_CanWithdraw(t *testing.T) {
	tests := []struct {
		name    string
		balance Balance
		amount  float64
		want    bool
	}{
		{
			name:    "can withdraw when sufficient funds",
			balance: Balance{Current: 100.0, Withdrawn: 50.0},
			amount:  50.0,
			want:    true,
		},
		{
			name:    "can withdraw exact amount",
			balance: Balance{Current: 100.0, Withdrawn: 0},
			amount:  100.0,
			want:    true,
		},
		{
			name:    "cannot withdraw when insufficient funds",
			balance: Balance{Current: 50.0, Withdrawn: 0},
			amount:  100.0,
			want:    false,
		},
		{
			name:    "can withdraw zero",
			balance: Balance{Current: 100.0, Withdrawn: 0},
			amount:  0.0,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.balance.CanWithdraw(tt.amount))
		})
	}
}

func TestWithdrawal_ToResponse(t *testing.T) {
	now := time.Now()

	withdrawal := Withdrawal{
		ID:          1,
		UserID:      2,
		OrderNumber: "2377225624",
		Sum:         500.0,
		ProcessedAt: now,
	}

	response := withdrawal.ToResponse()

	assert.Equal(t, withdrawal.OrderNumber, response.Order)
	assert.Equal(t, withdrawal.Sum, response.Sum)
	assert.Equal(t, withdrawal.ProcessedAt, response.ProcessedAt)
}

func TestWithdrawRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     WithdrawRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: WithdrawRequest{
				Order: "2377225624",
				Sum:   100.0,
			},
			wantErr: false,
		},
		{
			name: "empty order",
			req: WithdrawRequest{
				Order: "",
				Sum:   100.0,
			},
			wantErr: true,
			errMsg:  "order number is required",
		},
		{
			name: "zero sum",
			req: WithdrawRequest{
				Order: "2377225624",
				Sum:   0,
			},
			wantErr: true,
			errMsg:  "sum must be positive",
		},
		{
			name: "negative sum",
			req: WithdrawRequest{
				Order: "2377225624",
				Sum:   -100.0,
			},
			wantErr: true,
			errMsg:  "sum must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
