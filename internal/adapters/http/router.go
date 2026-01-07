// Package http provides HTTP server components.
package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/avitamin/go-gophermart/internal/adapters/http/handler"
	"github.com/avitamin/go-gophermart/internal/adapters/http/middleware"
	"github.com/avitamin/go-gophermart/internal/pkg/auth"
)

// Router holds all HTTP handlers.
type Router struct {
	UserHandler    *handler.UserHandler
	OrderHandler   *handler.OrderHandler
	BalanceHandler *handler.BalanceHandler
	JWTManager     *auth.JWTManager
}

// NewRouter creates a new chi router with all routes configured.
func NewRouter(r *Router) chi.Router {
	router := chi.NewRouter()

	// Global middleware
	router.Use(middleware.Logging)
	router.Use(middleware.Gzip)

	// Public routes
	router.Post("/api/user/register", r.UserHandler.Register)
	router.Post("/api/user/login", r.UserHandler.Login)

	// Protected routes
	router.Group(func(router chi.Router) {
		router.Use(middleware.Auth(r.JWTManager))

		router.Post("/api/user/orders", r.OrderHandler.CreateOrder)
		router.Get("/api/user/orders", r.OrderHandler.GetOrders)
		router.Get("/api/user/balance", r.BalanceHandler.GetBalance)
		router.Post("/api/user/balance/withdraw", r.BalanceHandler.Withdraw)
		router.Get("/api/user/withdrawals", r.BalanceHandler.GetWithdrawals)
	})

	return router
}
