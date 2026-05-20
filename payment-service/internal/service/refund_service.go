package service

import (
	"context"
	"errors"
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
}

type RefundService struct {
	refundRepo RefundRepository
	orderRepo  interface{}
	creditSvc  interface{}
}

func NewRefundService(refundRepo RefundRepository, orderRepo, creditSvc interface{}) *RefundService {
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
	return s.refundRepo.UpdateStatus(ctx, refundID, "approved", approverID, "")
}

func (s *RefundService) RejectRefund(ctx context.Context, refundID, approverID int64, note string) error {
	return s.refundRepo.UpdateStatus(ctx, refundID, "rejected", approverID, note)
}
