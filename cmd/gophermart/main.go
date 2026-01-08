// Package main is the entry point for the Gophermart loyalty system.
package main

import (
	"context"
	"errors"
	"net/http"
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
	"golang.org/x/sync/errgroup"
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
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Create HTTP server
	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}

	// Start accrual worker if configured
	var accrualWorker *accrual.Worker
	if cfg.AccrualSystemAddress != "" {
		accrualClient := accrual.NewClient(cfg.AccrualSystemAddress, log)
		accrualWorker = accrual.NewWorker(accrualClient, orderService, cfg.WorkerCount, cfg.WorkerInterval, log)
		accrualWorker.Start(ctx)
	}

	// Use errgroup for coordinated server lifecycle
	eg, egCtx := errgroup.WithContext(ctx)

	// Run HTTP server
	eg.Go(func() error {
		log.Info("starting server", zap.String("address", cfg.RunAddress))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	// Handle graceful shutdown
	eg.Go(func() error {
		<-egCtx.Done()

		log.Info("shutting down server...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("server shutdown error", zap.Error(err))
		}

		if accrualWorker != nil {
			accrualWorker.Stop()
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Fatal("server error", zap.Error(err))
	}

	log.Info("server stopped")
}
