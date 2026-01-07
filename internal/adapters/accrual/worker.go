// Package accrual provides HTTP client and worker for accrual system.
package accrual

import (
	"context"
	"sync"
	"time"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/avitamin/go-gophermart/internal/domain/service"
	"github.com/avitamin/go-gophermart/internal/pkg/logger"
	"go.uber.org/zap"
)

// Worker processes orders by fetching their accrual status.
type Worker struct {
	client       *Client
	orderService *service.OrderService
	workerCount  int
	interval     time.Duration
	orderChan    chan entity.Order
	wg           sync.WaitGroup
	rateLimitMu  sync.RWMutex
	rateLimitEnd time.Time
}

// NewWorker creates a new accrual worker pool.
func NewWorker(client *Client, orderService *service.OrderService, workerCount int, interval time.Duration) *Worker {
	return &Worker{
		client:       client,
		orderService: orderService,
		workerCount:  workerCount,
		interval:     interval,
		orderChan:    make(chan entity.Order, 100),
	}
}

// Start starts the worker pool.
func (w *Worker) Start(ctx context.Context) {
	logger.Info("starting accrual workers", zap.Int("count", w.workerCount))

	// Start workers
	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.worker(ctx, i)
	}

	// Start order fetcher
	w.wg.Add(1)
	go w.fetchOrders(ctx)
}

// Stop waits for all workers to finish.
func (w *Worker) Stop() {
	close(w.orderChan)
	w.wg.Wait()
	logger.Info("accrual workers stopped")
}

func (w *Worker) fetchOrders(ctx context.Context) {
	defer w.wg.Done()
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.loadPendingOrders(ctx)
		}
	}
}

func (w *Worker) loadPendingOrders(ctx context.Context) {
	// Check rate limit
	w.rateLimitMu.RLock()
	if time.Now().Before(w.rateLimitEnd) {
		w.rateLimitMu.RUnlock()
		return
	}
	w.rateLimitMu.RUnlock()

	orders, err := w.orderService.GetPendingOrders(ctx)
	if err != nil {
		logger.Error("failed to get pending orders", zap.Error(err))
		return
	}

	for _, order := range orders {
		select {
		case <-ctx.Done():
			return
		case w.orderChan <- order:
		default:
			// Channel full, skip this order for now
		}
	}
}

func (w *Worker) worker(ctx context.Context, id int) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case order, ok := <-w.orderChan:
			if !ok {
				return
			}
			w.processOrder(ctx, order)
		}
	}
}

func (w *Worker) processOrder(ctx context.Context, order entity.Order) {
	// Check rate limit
	w.rateLimitMu.RLock()
	if time.Now().Before(w.rateLimitEnd) {
		w.rateLimitMu.RUnlock()
		time.Sleep(time.Until(w.rateLimitEnd))
	} else {
		w.rateLimitMu.RUnlock()
	}

	accrualResp, err := w.client.GetOrderAccrual(ctx, order.Number)
	if err != nil {
		if rateLimitErr, ok := IsRateLimited(err); ok {
			logger.Warn("rate limited by accrual service",
				zap.String("order", order.Number),
				zap.Duration("retry_after", rateLimitErr.RetryAfter),
			)
			w.rateLimitMu.Lock()
			w.rateLimitEnd = time.Now().Add(rateLimitErr.RetryAfter)
			w.rateLimitMu.Unlock()
			return
		}
		logger.Error("failed to get order accrual",
			zap.String("order", order.Number),
			zap.Error(err),
		)
		return
	}

	if accrualResp == nil {
		// Order not registered in accrual system
		return
	}

	// Update order status if changed
	if accrualResp.Status != order.Status {
		if err := w.orderService.UpdateOrderStatus(ctx, order.Number, accrualResp.Status, accrualResp.Accrual); err != nil {
			logger.Error("failed to update order status",
				zap.String("order", order.Number),
				zap.Error(err),
			)
			return
		}
		logger.Info("order status updated",
			zap.String("order", order.Number),
			zap.String("status", string(accrualResp.Status)),
		)
	}
}
