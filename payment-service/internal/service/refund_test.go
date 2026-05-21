package service

import (
	"context"
	"testing"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

func TestCalculateRefundFullWithin7Days(t *testing.T) {
	repo := &mockRefundRepo{}
	svc := NewRefundService(repo, nil, nil)
	order := &model.Order{
		ID:        1,
		UserID:    42,
		OrderNo:   "ORD001",
		Amount:    100,
		Status:    model.OrderStatusPaid,
		CreatedAt: time.Now().Add(-2 * 24 * time.Hour),
	}
	refund, err := svc.CalculateRefund(context.Background(), order)
	if err != nil {
		t.Fatalf("CalculateRefund failed: %v", err)
	}
	if refund.Amount != 100 {
		t.Fatalf("expected full refund 100, got %.2f", refund.Amount)
	}
}

func TestCalculateRefundProratedAfter7Days(t *testing.T) {
	svc := NewRefundService(nil, nil, nil)
	order := &model.Order{
		ID:        2,
		UserID:    42,
		OrderNo:   "ORD002",
		Amount:    300,
		Status:    model.OrderStatusPaid,
		CreatedAt: time.Now().Add(-20 * 24 * time.Hour),
	}
	refund, err := svc.CalculateRefund(context.Background(), order)
	if err != nil {
		t.Fatalf("CalculateRefund failed: %v", err)
	}
	if refund.Amount >= 300 {
		t.Fatalf("expected prorated refund < 300, got %.2f", refund.Amount)
	}
	if refund.Amount <= 0 {
		t.Fatalf("expected positive refund amount, got %.2f", refund.Amount)
	}
}

func TestRequestRefundRejectsNonPaidOrder(t *testing.T) {
	repo := &mockRefundRepo{}
	orderRepo := &mockRefundOrderRepo{order: &model.Order{
		ID:     1,
		Status: model.OrderStatusPending,
	}}
	svc := NewRefundService(repo, orderRepo, nil)
	_, err := svc.RequestRefund(context.Background(), 1, 42, "test")
	if err != ErrOrderNotPaid {
		t.Fatalf("expected ErrOrderNotPaid, got %v", err)
	}
}

func TestRefundStatusFlow(t *testing.T) {
	repo := &mockRefundRepo{}
	orderRepo := &mockRefundOrderRepo{order: &model.Order{
		ID:        1,
		UserID:    42,
		Status:    model.OrderStatusPaid,
		Amount:    100,
		CreatedAt: time.Now().Add(-2 * 24 * time.Hour),
	}}
	svc := NewRefundService(repo, orderRepo, nil)
	refund, err := svc.RequestRefund(context.Background(), 1, 42, "产品质量问题")
	if err != nil {
		t.Fatalf("RequestRefund failed: %v", err)
	}
	if refund.Status != "pending" {
		t.Fatalf("expected pending status, got %s", refund.Status)
	}
	err = svc.ApproveRefund(context.Background(), refund.ID, 1)
	if err != nil {
		t.Fatalf("ApproveRefund failed: %v", err)
	}
}

type mockRefundOrderRepo struct {
	order *model.Order
}

func (m *mockRefundOrderRepo) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	if m.order != nil {
		return m.order, nil
	}
	return nil, nil
}

func (m *mockRefundOrderRepo) GetPendingOrdersOlderThan(ctx context.Context, since time.Duration) ([]*model.Order, error) {
	return nil, nil
}

func (m *mockRefundOrderRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	return nil
}

func (m *mockRefundOrderRepo) UpdateOrderStatus(ctx context.Context, orderNo string, fromStatus, toStatus string) error {
	return nil
}

type mockRefundRepo struct{}

func (m *mockRefundRepo) FindByOrderID(ctx context.Context, orderID int64) (*model.Refund, error) {
	return nil, nil
}

func (m *mockRefundRepo) Create(ctx context.Context, r *model.Refund) error {
	r.ID = 1
	return nil
}

func (m *mockRefundRepo) GetByID(ctx context.Context, id int64) (*model.Refund, error) {
	return &model.Refund{ID: id, Status: "pending", OrderID: 1, Amount: 100, UserID: 42}, nil
}

func (m *mockRefundRepo) UpdateStatus(ctx context.Context, id int64, status string, approverID int64, note string) error {
	return nil
}
