// Package accrual provides HTTP client for accrual system.
package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/avitamin/go-gophermart/internal/domain/repository"
	"github.com/avitamin/go-gophermart/internal/pkg/logger"
	"go.uber.org/zap"
)

// ErrRateLimited is returned when accrual system returns 429.
type ErrRateLimited struct {
	RetryAfter time.Duration
}

func (e *ErrRateLimited) Error() string {
	return fmt.Sprintf("rate limited, retry after %v", e.RetryAfter)
}

// Client implements repository.AccrualClient.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new accrual system client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// accrualResponse represents the JSON response from accrual system.
type accrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

// GetOrderAccrual retrieves accrual information for an order.
func (c *Client) GetOrderAccrual(ctx context.Context, orderNumber string) (*repository.AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp accrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return nil, err
		}

		result := &repository.AccrualResponse{
			Order:  accrualResp.Order,
			Status: mapStatus(accrualResp.Status),
		}

		if accrualResp.Accrual > 0 {
			result.Accrual = &accrualResp.Accrual
		}

		return result, nil

	case http.StatusNoContent:
		return nil, nil

	case http.StatusTooManyRequests:
		retryAfter := 60 * time.Second
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if val, err := strconv.Atoi(ra); err == nil {
				retryAfter = time.Duration(val) * time.Second
			}
		}
		return nil, &ErrRateLimited{RetryAfter: retryAfter}

	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func mapStatus(status string) entity.OrderStatus {
	switch status {
	case "REGISTERED":
		return entity.OrderStatusNew
	case "PROCESSING":
		return entity.OrderStatusProcessing
	case "INVALID":
		return entity.OrderStatusInvalid
	case "PROCESSED":
		return entity.OrderStatusProcessed
	default:
		logger.Warn("unknown accrual status", zap.String("status", status))
		return entity.OrderStatusNew
	}
}

// IsRateLimited checks if error is rate limit error.
func IsRateLimited(err error) (*ErrRateLimited, bool) {
	var rateLimitErr *ErrRateLimited
	if errors.As(err, &rateLimitErr) {
		return rateLimitErr, true
	}
	return nil, false
}
