// Package accrual provides HTTP client and worker for accrual system.
package accrual

import (
	"context"
	"sync"
	"time"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/avitamin/go-gophermart/internal/domain/service"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Worker processes orders by fetching their accrual status.
type Worker struct {
	client       *Client
	orderService *service.OrderService
	workerCount  int
	interval     time.Duration
	orderChan    chan entity.Order
	eg           *errgroup.Group
	egCtx        context.Context
	cancel       context.CancelFunc
	rateLimitMu  sync.RWMutex
	rateLimitEnd time.Time
	log          *zap.Logger
}

// NewWorker creates a new accrual worker pool.
func NewWorker(client *Client, orderService *service.OrderService, workerCount int, interval time.Duration, log *zap.Logger) *Worker {
	return &Worker{
		client:       client,
		orderService: orderService,
		workerCount:  workerCount,
		interval:     interval,
		orderChan:    make(chan entity.Order, 100),
		log:          log,
	}
}

// Start starts the worker pool.
func (w *Worker) Start(ctx context.Context) {
	w.log.Info("starting accrual workers", zap.Int("count", w.workerCount))

	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	w.eg, w.egCtx = errgroup.WithContext(ctx)

	// Start workers
	for i := 0; i < w.workerCount; i++ {
		w.eg.Go(w.workerFunc(i))
	}

	// Start order fetcher
	w.eg.Go(w.fetchOrdersFunc())
}

// Stop stops all workers and waits for them to finish.
func (w *Worker) Stop() {
	w.cancel()
	close(w.orderChan)
	if err := w.eg.Wait(); err != nil {
		w.log.Error("worker pool error", zap.Error(err))
	}
	w.log.Info("accrual workers stopped")
}

func (w *Worker) fetchOrdersFunc() func() error {
	return func() error {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-w.egCtx.Done():
				return nil
			case <-ticker.C:
				w.loadPendingOrders(w.egCtx)
			}
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
		w.log.Error("failed to get pending orders", zap.Error(err))
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

func (w *Worker) workerFunc(id int) func() error {
	return func() error {
		for {
			select {
			case <-w.egCtx.Done():
				return nil
			case order, ok := <-w.orderChan:
				if !ok {
					return nil
				}
				w.processOrder(w.egCtx, order)
			}
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
			w.log.Warn("rate limited by accrual service",
				zap.String("order", order.Number),
				zap.Duration("retry_after", rateLimitErr.RetryAfter),
			)
			w.rateLimitMu.Lock()
			w.rateLimitEnd = time.Now().Add(rateLimitErr.RetryAfter)
			w.rateLimitMu.Unlock()
			return
		}
		w.log.Error("failed to get order accrual",
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
			w.log.Error("failed to update order status",
				zap.String("order", order.Number),
				zap.Error(err),
			)
			return
		}
		w.log.Info("order status updated",
			zap.String("order", order.Number),
			zap.String("status", string(accrualResp.Status)),
		)
	}
}
