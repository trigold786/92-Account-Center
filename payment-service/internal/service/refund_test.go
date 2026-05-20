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

func TestRefundStatusFlow(t *testing.T) {
	repo := &mockRefundRepo{}
	svc := NewRefundService(repo, nil, nil)
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

type mockRefundRepo struct{}

func (m *mockRefundRepo) Create(ctx context.Context, r *model.Refund) error {
	r.ID = 1
	return nil
}

func (m *mockRefundRepo) GetByID(ctx context.Context, id int64) (*model.Refund, error) {
	return &model.Refund{ID: id, Status: "pending", OrderID: 1, Amount: 100}, nil
}

func (m *mockRefundRepo) UpdateStatus(ctx context.Context, id int64, status string, approverID int64, note string) error {
	return nil
}
