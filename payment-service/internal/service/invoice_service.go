package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

type InvoiceRepository interface {
	Create(ctx context.Context, inv *model.Invoice) error
	GetByID(ctx context.Context, id int64) (*model.Invoice, error)
	GetByUserID(ctx context.Context, userID int64) ([]*model.Invoice, error)
}

type InvoiceService struct {
	repo      InvoiceRepository
	invoices  []*model.Invoice
	mu        sync.RWMutex
	idCounter atomic.Int64
}

func NewInvoiceService(repo InvoiceRepository) *InvoiceService {
	return &InvoiceService{repo: repo}
}

func (s *InvoiceService) CreateInvoice(ctx context.Context, inv *model.Invoice) (*model.Invoice, error) {
	if inv.Status == "" {
		inv.Status = "pending"
	}
	inv.CreatedAt = time.Now()
	inv.UpdatedAt = inv.CreatedAt

	if s.repo != nil {
		if err := s.repo.Create(ctx, inv); err != nil {
			return nil, fmt.Errorf("failed to create invoice: %w", err)
		}
		return inv, nil
	}

	inv.ID = s.idCounter.Add(1)

	s.mu.Lock()
	s.invoices = append(s.invoices, inv)
	s.mu.Unlock()
	return inv, nil
}

func (s *InvoiceService) GetInvoice(ctx context.Context, id int64) (*model.Invoice, error) {
	if s.repo != nil {
		return s.repo.GetByID(ctx, id)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, inv := range s.invoices {
		if inv.ID == id {
			return inv, nil
		}
	}
	return nil, fmt.Errorf("invoice not found")
}

func (s *InvoiceService) ListInvoices(ctx context.Context, userID int64) ([]*model.Invoice, error) {
	if s.repo != nil {
		return s.repo.GetByUserID(ctx, userID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.Invoice
	for _, inv := range s.invoices {
		if inv.UserID == userID {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (s *InvoiceService) AutoInvoice(ctx context.Context, orderID int64, userID int64, amount float64) (*model.Invoice, error) {
	inv := &model.Invoice{
		UserID:    userID,
		OrderID:   orderID,
		InvoiceNo: fmt.Sprintf("INV-AUTO-%d-%d", orderID, time.Now().Unix()),
		Title:     fmt.Sprintf("Invoice for Order #%d", orderID),
		Amount:    amount,
		Status:    "auto_generated",
	}

	return s.CreateInvoice(ctx, inv)
}
