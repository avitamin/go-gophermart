// Package handler provides HTTP handlers for the application.
package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/avitamin/go-gophermart/internal/adapters/http/middleware"
	"github.com/avitamin/go-gophermart/internal/domain/service"
	"github.com/avitamin/go-gophermart/internal/pkg/logger"
	"go.uber.org/zap"
)

// OrderHandler handles order-related HTTP requests.
type OrderHandler struct {
	orderService *service.OrderService
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// CreateOrder handles order creation.
// POST /api/user/orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		http.Error(w, "Order number is required", http.StatusBadRequest)
		return
	}

	err = h.orderService.CreateOrder(r.Context(), userID, orderNumber)
	if err != nil {
		if service.OrderExistsForUser(err) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if service.OrderBelongsToOther(err) {
			http.Error(w, "Order belongs to another user", http.StatusConflict)
			return
		}
		if err == service.ErrInvalidOrderNumber {
			http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
			return
		}
		logger.Error("create order failed", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetOrders handles order list retrieval.
// GET /api/user/orders
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.orderService.GetUserOrders(r.Context(), userID)
	if err != nil {
		logger.Error("get orders failed", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Convert to response format
	response := make([]map[string]interface{}, 0, len(orders))
	for _, order := range orders {
		item := map[string]interface{}{
			"number":      order.Number,
			"status":      order.Status,
			"uploaded_at": order.UploadedAt,
		}
		if order.Accrual != nil {
			item["accrual"] = *order.Accrual
		}
		response = append(response, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
