// Package http provides HTTP server components.
package http

import (
	"net/http"

	"github.com/avitamin/go-gophermart/internal/adapters/http/handler"
	"github.com/avitamin/go-gophermart/internal/adapters/http/middleware"
	"github.com/avitamin/go-gophermart/internal/pkg/auth"
	"go.uber.org/zap"
)

// Router holds all HTTP handlers.
type Router struct {
	UserHandler    *handler.UserHandler
	OrderHandler   *handler.OrderHandler
	BalanceHandler *handler.BalanceHandler
	JWTManager     *auth.JWTManager
	Logger         *zap.Logger
}

// NewRouter creates a new router with all routes configured.
func NewRouter(r *Router) http.Handler {
	mux := http.NewServeMux()

	auth := middleware.Auth(r.JWTManager)

	// Public routes
	mux.HandleFunc("POST /api/user/register", r.UserHandler.Register)
	mux.HandleFunc("POST /api/user/login", r.UserHandler.Login)

	// Protected routes
	mux.Handle("POST /api/user/orders", auth(http.HandlerFunc(r.OrderHandler.CreateOrder)))
	mux.Handle("GET /api/user/orders", auth(http.HandlerFunc(r.OrderHandler.GetOrders)))
	mux.Handle("GET /api/user/balance", auth(http.HandlerFunc(r.BalanceHandler.GetBalance)))
	mux.Handle("POST /api/user/balance/withdraw", auth(http.HandlerFunc(r.BalanceHandler.Withdraw)))
	mux.Handle("GET /api/user/withdrawals", auth(http.HandlerFunc(r.BalanceHandler.GetWithdrawals)))

	// Apply global middleware (order: Logging -> Gzip -> mux)
	return chain(mux, middleware.Gzip, middleware.Logging(r.Logger))
}

// chain applies middlewares in reverse order (last middleware wraps first).
func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}
