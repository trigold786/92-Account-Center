package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
	"github.com/trigold786/92-Account-Center/payment-service/internal/provider"
	"github.com/trigold786/92-Account-Center/payment-service/internal/repository"
)

type ReconciliationService interface {
	ReconcileOrders(ctx context.Context, providerName string, date string) (*ReconciliationReport, error)
	GetReconciliationReport(ctx context.Context, reportID string) (*ReconciliationReport, error)
}

type ReconciliationReportRepository interface {
	Save(ctx context.Context, report *ReconciliationReport) error
	GetByID(ctx context.Context, id string) (*ReconciliationReport, error)
}

type ReconciliationReport struct {
	ID             string          `json:"id"`
	ProviderName   string          `json:"provider_name"`
	Date           string          `json:"date"`
	TotalOrders    int             `json:"total_orders"`
	MatchedOrders  int             `json:"matched_orders"`
	MismatchOrders []MismatchOrder `json:"mismatch_orders,omitempty"`
	Status         string          `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
}

type MismatchOrder struct {
	OrderNo      string  `json:"order_no"`
	LocalStatus  string  `json:"local_status"`
	RemoteStatus string  `json:"remote_status"`
	LocalAmount  float64 `json:"local_amount"`
	RemoteAmount float64 `json:"remote_amount"`
}

type reconciliationService struct {
	providerRegistry *provider.ProviderRegistry
	orderRepo        repository.OrderRepository
	logger           *slog.Logger
	reportRepo       ReconciliationReportRepository
	mu               sync.RWMutex
	reports          map[string]*ReconciliationReport
}

func NewReconciliationService(
	providerRegistry *provider.ProviderRegistry,
	orderRepo repository.OrderRepository,
	logger *slog.Logger,
) ReconciliationService {
	return NewReconciliationServiceWithRepository(providerRegistry, orderRepo, nil, logger)
}

func NewReconciliationServiceWithRepository(
	providerRegistry *provider.ProviderRegistry,
	orderRepo repository.OrderRepository,
	reportRepo ReconciliationReportRepository,
	logger *slog.Logger,
) ReconciliationService {
	return &reconciliationService{
		providerRegistry: providerRegistry,
		orderRepo:        orderRepo,
		logger:           logger,
		reportRepo:       reportRepo,
		reports:          make(map[string]*ReconciliationReport),
	}
}

func (s *reconciliationService) ReconcileOrders(ctx context.Context, providerName string, date string) (*ReconciliationReport, error) {
	p, ok := s.providerRegistry.Get(providerName)
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	startTime := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, parsedDate.Location())
	endTime := startTime.Add(24 * time.Hour)

	paidStatus := model.OrderStatusPaid
	query := &model.OrderQueryRequest{
		Status:    &paidStatus,
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      1,
		PageSize:  1000,
	}

	orders, total, err := s.orderRepo.List(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}

	report := &ReconciliationReport{
		ID:             fmt.Sprintf("RECON-%s-%s", providerName, date),
		ProviderName:   providerName,
		Date:           date,
		TotalOrders:    total,
		MatchedOrders:  0,
		MismatchOrders: []MismatchOrder{},
		Status:         "completed",
		CreatedAt:      time.Now(),
	}

	for _, order := range orders {
		remoteStatus, err := p.QueryPayment(ctx, order.OrderNo)
		if err != nil {
			s.logger.Error("failed to query remote payment status",
				"order_no", order.OrderNo, "error", err)
			report.Status = "partial"
			continue
		}

		localStatus := mapLocalStatus(order.Status)
		matched := true

		if localStatus != remoteStatus.Status {
			matched = false
		}
		if order.Amount != remoteStatus.Amount && remoteStatus.Amount > 0 {
			matched = false
		}

		if matched {
			report.MatchedOrders++
		} else {
			report.MismatchOrders = append(report.MismatchOrders, MismatchOrder{
				OrderNo:      order.OrderNo,
				LocalStatus:  string(order.Status),
				RemoteStatus: remoteStatus.Status,
				LocalAmount:  order.Amount,
				RemoteAmount: remoteStatus.Amount,
			})
		}
	}

	s.mu.Lock()
	s.reports[report.ID] = report
	s.mu.Unlock()
	if s.reportRepo != nil {
		if err := s.reportRepo.Save(ctx, report); err != nil {
			return nil, fmt.Errorf("failed to persist reconciliation report: %w", err)
		}
	}

	return report, nil
}

func (s *reconciliationService) GetReconciliationReport(ctx context.Context, reportID string) (*ReconciliationReport, error) {
	s.mu.RLock()
	report, ok := s.reports[reportID]
	s.mu.RUnlock()
	if ok {
		return report, nil
	}
	if s.reportRepo != nil {
		report, err := s.reportRepo.GetByID(ctx, reportID)
		if err == nil && report != nil {
			return report, nil
		}
	}
	return nil, fmt.Errorf("report not found: %s", reportID)
}

func mapLocalStatus(status model.OrderStatus) string {
	switch status {
	case model.OrderStatusPaid:
		return "SUCCESS"
	case model.OrderStatusPending:
		return "NOTPAY"
	case model.OrderStatusCancelled:
		return "CLOSED"
	case model.OrderStatusRefunded:
		return "REFUND"
	default:
		return string(status)
	}
}
