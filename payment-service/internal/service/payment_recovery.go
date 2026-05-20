package service

import (
	"context"
	"log"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

type OrderRepository interface {
	GetPendingOrdersOlderThan(ctx context.Context, since time.Duration) ([]*model.Order, error)
	UpdateOrderStatus(ctx context.Context, orderNo string, fromStatus, toStatus string) error
}

type PaymentRecoveryService struct {
	orderRepo OrderRepository
	providers interface{}
}

func NewPaymentRecoveryService(orderRepo OrderRepository, providers interface{}) *PaymentRecoveryService {
	return &PaymentRecoveryService{orderRepo: orderRepo, providers: providers}
}

func (s *PaymentRecoveryService) RecoverPendingOrders(ctx context.Context, since time.Duration) (int, error) {
	orders, err := s.orderRepo.GetPendingOrdersOlderThan(ctx, since)
	if err != nil {
		return 0, err
	}
	recovered := 0
	for _, o := range orders {
		log.Printf("Recovering pending order %s (amount=%.2f)", o.OrderNo, o.Amount)
		if err := s.orderRepo.UpdateOrderStatus(ctx, o.OrderNo, string(model.OrderStatusPending), string(model.OrderStatusCancelled)); err != nil {
			log.Printf("Failed to recover order %s: %v", o.OrderNo, err)
			continue
		}
		recovered++
	}
	return recovered, nil
}

func (s *PaymentRecoveryService) RunScheduledRecovery(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			count, err := s.RecoverPendingOrders(ctx, 5*time.Minute)
			if err != nil {
				log.Printf("Scheduled recovery error: %v", err)
				continue
			}
			if count > 0 {
				log.Printf("Recovered %d stale pending orders", count)
			}
		case <-ctx.Done():
			return
		}
	}
}
