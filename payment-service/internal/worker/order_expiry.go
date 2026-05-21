package worker

import (
	"context"
	"log"
	"time"
)

type OrderRepository interface {
	FindExpired(ctx context.Context, before time.Time) ([]Order, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
}

type Order struct {
	ID     int64
	Status string
}

type OrderExpiryWorker struct {
	orderRepo OrderRepository
	interval  time.Duration
	done      chan struct{}
}

func NewOrderExpiryWorker(orderRepo OrderRepository, interval time.Duration) *OrderExpiryWorker {
	if interval == 0 {
		interval = 5 * time.Minute
	}
	return &OrderExpiryWorker{
		orderRepo: orderRepo,
		interval:  interval,
		done:      make(chan struct{}),
	}
}

func (w *OrderExpiryWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.done:
			return
		case <-ticker.C:
			w.processExpiredOrders(ctx)
		}
	}
}

func (w *OrderExpiryWorker) Stop() {
	close(w.done)
}

func (w *OrderExpiryWorker) processExpiredOrders(ctx context.Context) {
	now := time.Now()
	orders, err := w.orderRepo.FindExpired(ctx, now)
	if err != nil {
		log.Printf("[OrderExpiryWorker] failed to find expired orders: %v", err)
		return
	}

	for _, order := range orders {
		if order.Status != "pending" {
			continue
		}
		if err := w.orderRepo.UpdateStatus(ctx, order.ID, "expired"); err != nil {
			log.Printf("[OrderExpiryWorker] failed to expire order %d: %v", order.ID, err)
			continue
		}
		log.Printf("[OrderExpiryWorker] order %d expired", order.ID)
	}
}
