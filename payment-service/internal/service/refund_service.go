package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

var (
	ErrOrderNotPaid        = errors.New("order is not paid")
	ErrRefundAlreadyExists = errors.New("refund already exists for this order")
)

type RefundRepository interface {
	Create(ctx context.Context, refund *model.Refund) error
	GetByID(ctx context.Context, id int64) (*model.Refund, error)
	UpdateStatus(ctx context.Context, id int64, status string, approverID int64, note string) error
	FindByOrderID(ctx context.Context, orderID int64) (*model.Refund, error)
}

type CreditService interface {
	ReverseCredits(ctx context.Context, userID int64, amount int, reason string) error
}

type RefundService struct {
	refundRepo RefundRepository
	orderRepo  OrderRepository
	creditSvc  CreditService
}

func NewRefundService(refundRepo RefundRepository, orderRepo OrderRepository, creditSvc CreditService) *RefundService {
	return &RefundService{refundRepo: refundRepo, orderRepo: orderRepo, creditSvc: creditSvc}
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

func (s *RefundService) ApproveRefund(ctx context.Context, refundID, approverID int64) error {
	refund, err := s.refundRepo.GetByID(ctx, refundID)
	if err != nil {
		return fmt.Errorf("failed to fetch refund: %w", err)
	}

	if err := s.refundRepo.UpdateStatus(ctx, refundID, "approved", approverID, ""); err != nil {
		return err
	}

	order, err := s.orderRepo.GetByID(ctx, refund.OrderID)
	if err != nil {
		_ = s.refundRepo.UpdateStatus(ctx, refundID, "pending", approverID, "")
		return fmt.Errorf("failed to fetch order for credit reversal: %w", err)
	}

	calculated, err := s.CalculateRefund(ctx, order)
	if err != nil {
		_ = s.refundRepo.UpdateStatus(ctx, refundID, "pending", approverID, "")
		return fmt.Errorf("failed to calculate refund: %w", err)
	}

	if s.creditSvc != nil {
		if err := s.creditSvc.ReverseCredits(ctx, order.UserID, int(calculated.Amount), fmt.Sprintf("refund for order %d", order.ID)); err != nil {
			_ = s.refundRepo.UpdateStatus(ctx, refundID, "pending", approverID, "")
			return fmt.Errorf("credit reversal failed, refund reverted to pending: %w", err)
		}
	}

	return nil
}

func (s *RefundService) RejectRefund(ctx context.Context, refundID, approverID int64, note string) error {
	return s.refundRepo.UpdateStatus(ctx, refundID, "rejected", approverID, note)
}
