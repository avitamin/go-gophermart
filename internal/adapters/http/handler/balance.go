// Package handler provides HTTP handlers for the application.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/avitamin/go-gophermart/internal/adapters/http/middleware"
	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/avitamin/go-gophermart/internal/domain/service"
	"go.uber.org/zap"
)

// BalanceHandler handles balance-related HTTP requests.
type BalanceHandler struct {
	balanceService *service.BalanceService
	log            *zap.Logger
}

// NewBalanceHandler creates a new BalanceHandler.
func NewBalanceHandler(balanceService *service.BalanceService, log *zap.Logger) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
		log:            log,
	}
}

// GetBalance handles balance retrieval.
// GET /api/user/balance
func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		h.log.Error("get balance failed", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// Withdraw handles withdrawal request.
// POST /api/user/balance/withdraw
func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req entity.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.balanceService.Withdraw(r.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, entity.ErrInsufficientFunds) {
			http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
			return
		}
		if err == service.ErrInvalidOrderNumber {
			http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
			return
		}
		h.log.Error("withdrawal failed", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetWithdrawals handles withdrawals list retrieval.
// GET /api/user/withdrawals
func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.balanceService.GetWithdrawals(r.Context(), userID)
	if err != nil {
		h.log.Error("get withdrawals failed", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Convert to response format
	response := make([]entity.WithdrawalResponse, 0, len(withdrawals))
	for _, w := range withdrawals {
		response = append(response, w.ToResponse())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
