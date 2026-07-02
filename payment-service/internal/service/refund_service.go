package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
	"github.com/trigold786/92-Account-Center/payment-service/internal/provider"
)

var (
	ErrOrderNotPaid        = errors.New("order is not paid")
	ErrRefundAlreadyExists = errors.New("refund already exists for this order")
)

// RefundRepository defines persistence operations for refund records.
type RefundRepository interface {
	Create(ctx context.Context, refund *model.Refund) error
	GetByID(ctx context.Context, id int64) (*model.Refund, error)
	UpdateStatus(ctx context.Context, id int64, status string, approverID int64, note string) error
	FindByOrderID(ctx context.Context, orderID int64) (*model.Refund, error)
	UpdateProviderResult(ctx context.Context, id int64, refundNo string, providerName string, providerRefundID string, providerStatus string) error
	MarkProviderFailure(ctx context.Context, id int64, refundNo string, providerName string, providerError string) error
}

// PaymentProviderRegistry provides lookup for payment providers by name.
type PaymentProviderRegistry interface {
	Get(name string) (provider.PaymentProvider, bool)
}

// CreditService handles reversal of credits after a refund.
type CreditService interface {
	ReverseCredits(ctx context.Context, userID int64, amount int, reason string) error
}

// SubscriptionCancellationNotifier cancels subscriptions tied to a refunded order.
type SubscriptionCancellationNotifier interface {
	CancelRefundedOrderSubscription(ctx context.Context, userID int64, orderID int64, reason string) error
}

// RefundService orchestrates refund creation, approval, and provider integration.
type RefundService struct {
	refundRepo       RefundRepository
	orderRepo        OrderRepository
	creditSvc        CreditService
	subCancelSvc     SubscriptionCancellationNotifier
	providerRegistry PaymentProviderRegistry
}

// NewRefundService creates a RefundService with optional notifier and provider registry.
func NewRefundService(refundRepo RefundRepository, orderRepo OrderRepository, creditSvc CreditService, opts ...any) *RefundService {
	var notifier SubscriptionCancellationNotifier
	var registry PaymentProviderRegistry
	for _, opt := range opts {
		switch o := opt.(type) {
		case SubscriptionCancellationNotifier:
			notifier = o
		case PaymentProviderRegistry:
			registry = o
		case nil:
			// skip nil
		}
	}
	return &RefundService{refundRepo: refundRepo, orderRepo: orderRepo, creditSvc: creditSvc, subCancelSvc: notifier, providerRegistry: registry}
}

func (s *RefundService) CalculateRefund(ctx context.Context, order *model.Order) (*model.Refund, error) {
	daysSinceCreation := time.Since(order.CreatedAt).Hours() / 24
	amount := order.Amount
	if daysSinceCreation > 7 {
		subscriptionDays := 30.0
		usedDays := math.Min(daysSinceCreation, subscriptionDays)
		remainingDays := subscriptionDays - usedDays
		if remainingDays < 1 {
			remainingDays = 1
		}
		amount = math.Round(order.Amount*remainingDays/subscriptionDays*100) / 100
	}
	return &model.Refund{
		OrderID: order.ID,
		Amount:  amount,
		Status:  "calculated",
	}, nil
}

// RequestRefund creates a pending refund record for a paid order.
func (s *RefundService) RequestRefund(ctx context.Context, orderID, userID int64, reason string) (*model.Refund, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order: %w", err)
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	if order.Status != model.OrderStatusPaid {
		return nil, ErrOrderNotPaid
	}

	existing, err := s.refundRepo.FindByOrderID(ctx, orderID)
	if err == nil && existing != nil {
		return nil, ErrRefundAlreadyExists
	}

	refund := &model.Refund{
		OrderID: orderID,
		UserID:  userID,
		Amount:  0,
		Reason:  reason,
		Status:  "pending",
	}
	if err := s.refundRepo.Create(ctx, refund); err != nil {
		return nil, err
	}
	return refund, nil
}

// ApproveRefund processes a pending refund through the payment provider and updates related records.
func (s *RefundService) ApproveRefund(ctx context.Context, refundID, approverID int64) error {
	refund, err := s.refundRepo.GetByID(ctx, refundID)
	if err != nil {
		return fmt.Errorf("failed to fetch refund: %w", err)
	}

	if refund.Status != "pending" {
		return fmt.Errorf("refund cannot be approved from status %s", refund.Status)
	}

	order, err := s.orderRepo.GetByID(ctx, refund.OrderID)
	if err != nil {
		return fmt.Errorf("failed to fetch order: %w", err)
	}

	calculated, err := s.CalculateRefund(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to calculate refund: %w", err)
	}

	refundNo := refundNoFor(refundID)

	if s.providerRegistry != nil {
		providerName, ok := providerNameFromPaymentMethod(order.PaymentMethod)
		if !ok {
			if err := s.refundRepo.MarkProviderFailure(ctx, refundID, refundNo, order.PaymentMethod, fmt.Sprintf("unsupported payment method: %s", order.PaymentMethod)); err != nil {
				slog.Error("failed to mark provider failure", "refund_id", refundID, "error", err)
			}
			return fmt.Errorf("unsupported payment method for refund: %s", order.PaymentMethod)
		}

		paymentProvider, ok := s.providerRegistry.Get(providerName)
		if !ok {
			if err := s.refundRepo.MarkProviderFailure(ctx, refundID, refundNo, providerName, fmt.Sprintf("provider %s not registered", providerName)); err != nil {
				slog.Error("failed to mark provider failure", "refund_id", refundID, "error", err)
			}
			return fmt.Errorf("payment provider %s not available", providerName)
		}

		refundResp, err := paymentProvider.Refund(ctx, &provider.RefundRequest{
			OrderNo:      order.OrderNo,
			RefundNo:     refundNo,
			TotalAmount:  order.Amount,
			RefundAmount: calculated.Amount,
			Reason:       refund.Reason,
		})
		if err != nil {
			if err := s.refundRepo.MarkProviderFailure(ctx, refundID, refundNo, providerName, err.Error()); err != nil {
				slog.Error("failed to mark provider failure", "refund_id", refundID, "error", err)
			}
			return fmt.Errorf("provider refund failed: %w", err)
		}

		if err := s.refundRepo.UpdateProviderResult(ctx, refundID, refundNo, providerName, refundResp.RefundID, refundResp.Status); err != nil {
			return fmt.Errorf("failed to persist provider refund result: %w", err)
		}
	}

	if err := s.refundRepo.UpdateStatus(ctx, refundID, "approved", approverID, ""); err != nil {
		return fmt.Errorf("failed to mark refund approved: %w", err)
	}

	if err := s.orderRepo.UpdateStatus(ctx, order.ID, string(model.OrderStatusRefunded)); err != nil {
		return fmt.Errorf("failed to mark order refunded: %w", err)
	}

	if order.ProductType == "subscription" && s.subCancelSvc != nil {
		if err := s.subCancelSvc.CancelRefundedOrderSubscription(ctx, order.UserID, order.ID, "refund approved"); err != nil {
			return fmt.Errorf("subscription cancellation failed: %w", err)
		}
	}

	if s.creditSvc != nil {
		if err := s.creditSvc.ReverseCredits(ctx, order.UserID, int(calculated.Amount), fmt.Sprintf("refund for order %d", order.ID)); err != nil {
			return fmt.Errorf("credit reversal failed: %w", err)
		}
	}

	return nil
}

func (s *RefundService) RejectRefund(ctx context.Context, refundID, approverID int64, note string) error {
	return s.refundRepo.UpdateStatus(ctx, refundID, "rejected", approverID, note)
}

func providerNameFromPaymentMethod(method string) (string, bool) {
	m := strings.ToLower(strings.TrimSpace(method))
	switch {
	case strings.HasPrefix(m, "wechat"):
		return "wechat", true
	case strings.HasPrefix(m, "alipay"):
		return "alipay", true
	default:
		return "", false
	}
}

func refundNoFor(refundID int64) string {
	return fmt.Sprintf("RF%d", refundID)
}
