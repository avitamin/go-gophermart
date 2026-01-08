// Package main is the entry point for the Gophermart loyalty system.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/avitamin/go-gophermart/internal/adapters/accrual"
	httpAdapter "github.com/avitamin/go-gophermart/internal/adapters/http"
	"github.com/avitamin/go-gophermart/internal/adapters/http/handler"
	"github.com/avitamin/go-gophermart/internal/adapters/postgres"
	"github.com/avitamin/go-gophermart/internal/config"
	"github.com/avitamin/go-gophermart/internal/domain/service"
	"github.com/avitamin/go-gophermart/internal/pkg/auth"
	"github.com/avitamin/go-gophermart/internal/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Initialize configuration
	cfg := config.New()

	// Initialize logger
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	// Validate required configuration
	if cfg.DatabaseURI == "" {
		log.Fatal("DATABASE_URI is required")
	}

	// Initialize database
	db, err := postgres.New(cfg.DatabaseURI, log)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	balanceRepo := postgres.NewBalanceRepository(db)

	// Initialize auth components
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.TokenDuration)
	hasher := auth.NewBcryptHasher()

	// Initialize services
	userService := service.NewUserService(userRepo, jwtManager, hasher)
	orderService := service.NewOrderService(orderRepo)
	balanceService := service.NewBalanceService(balanceRepo)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService, log)
	orderHandler := handler.NewOrderHandler(orderService, log)
	balanceHandler := handler.NewBalanceHandler(balanceService, log)

	// Create router
	router := httpAdapter.NewRouter(&httpAdapter.Router{
		UserHandler:    userHandler,
		OrderHandler:   orderHandler,
		BalanceHandler: balanceHandler,
		JWTManager:     jwtManager,
		Logger:         log,
	})

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start accrual worker if configured
	var accrualWorker *accrual.Worker
	if cfg.AccrualSystemAddress != "" {
		accrualClient := accrual.NewClient(cfg.AccrualSystemAddress, log)
		accrualWorker = accrual.NewWorker(accrualClient, orderService, cfg.WorkerCount, cfg.WorkerInterval, log)
		accrualWorker.Start(ctx)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}

	// Handle shutdown signals
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Info("shutting down server...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("server shutdown error", zap.Error(err))
		}

		if accrualWorker != nil {
			accrualWorker.Stop()
		}
	}()

	// Start server
	log.Info("starting server", zap.String("address", cfg.RunAddress))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("server error", zap.Error(err))
	}

	log.Info("server stopped")
}
