package service

import (
	"context"
	"testing"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

func TestCreateInvoiceSvc(t *testing.T) {
	svc := NewInvoiceService(nil)

	inv, err := svc.CreateInvoice(context.Background(), &model.Invoice{
		UserID:    1,
		OrderID:   100,
		InvoiceNo: "INV-001",
		Title:     "Test Invoice",
		Amount:    99.99,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if inv.Status != "pending" {
		t.Fatalf("expected pending, got %s", inv.Status)
	}
}

func TestAutoInvoice(t *testing.T) {
	svc := NewInvoiceService(nil)

	inv, err := svc.AutoInvoice(context.Background(), 200, 1, 49.99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != "auto_generated" {
		t.Fatalf("expected auto_generated, got %s", inv.Status)
	}
	if inv.OrderID != 200 {
		t.Fatalf("expected order 200, got %d", inv.OrderID)
	}
	if inv.Amount != 49.99 {
		t.Fatalf("expected 49.99, got %f", inv.Amount)
	}

	list, err := svc.ListInvoices(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 invoice, got %d", len(list))
	}
}
