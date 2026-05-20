package service

import (
	"context"
	"testing"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

func TestRecoverPendingOrders(t *testing.T) {
	repo := &mockRecoveryOrderRepo{
		orders: []*model.Order{
			{ID: 1, OrderNo: "ORD001", Status: model.OrderStatusPending, Amount: 100, CreatedAt: time.Now().Add(-10 * time.Minute)},
			{ID: 2, OrderNo: "ORD002", Status: model.OrderStatusPending, Amount: 200, CreatedAt: time.Now().Add(-2 * time.Minute)},
		},
	}
	svc := NewPaymentRecoveryService(repo, nil)
	recovered, err := svc.RecoverPendingOrders(context.Background(), 5*time.Minute)
	if err != nil {
		t.Fatalf("RecoverPendingOrders failed: %v", err)
	}
	if recovered != 1 {
		t.Fatalf("expected 1 recovered order (older than 5min), got %d", recovered)
	}
}

func TestRecoverNoPending(t *testing.T) {
	repo := &mockRecoveryOrderRepo{
		orders: []*model.Order{
			{ID: 3, OrderNo: "ORD003", Status: model.OrderStatusPaid, Amount: 300},
		},
	}
	svc := NewPaymentRecoveryService(repo, nil)
	recovered, err := svc.RecoverPendingOrders(context.Background(), 5*time.Minute)
	if err != nil {
		t.Fatalf("RecoverPendingOrders failed: %v", err)
	}
	if recovered != 0 {
		t.Fatalf("expected 0 recovered, got %d", recovered)
	}
}

type mockRecoveryOrderRepo struct {
	orders []*model.Order
}

func (m *mockRecoveryOrderRepo) GetPendingOrdersOlderThan(ctx context.Context, since time.Duration) ([]*model.Order, error) {
	var pending []*model.Order
	cutoff := time.Now().Add(-since)
	for _, o := range m.orders {
		if o.Status == model.OrderStatusPending && o.CreatedAt.Before(cutoff) {
			pending = append(pending, o)
		}
	}
	return pending, nil
}

func (m *mockRecoveryOrderRepo) UpdateOrderStatus(ctx context.Context, orderNo string, fromStatus, toStatus string) error {
	return nil
}
