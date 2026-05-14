package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
	"github.com/trigold786/92-Account-Center/account-service/internal/repository"
)

var (
	ErrSubscriptionNotFound = errors.New("no active subscription found")
	ErrInvalidTierUpgrade   = errors.New("invalid tier upgrade")
	ErrAlreadySubscribed    = errors.New("user already has an active subscription at this or higher tier")
)

const defaultDuration = 30 * 24 * time.Hour

type SubscriptionService interface {
	PurchaseSubscription(ctx context.Context, req *model.PurchaseRequest) (*model.Subscription, error)
	UpgradeSubscription(ctx context.Context, req *model.UpgradeRequest) (*model.Subscription, error)
	RenewSubscription(ctx context.Context, req *model.RenewRequest) (*model.Subscription, error)
	GetUserSubscriptions(ctx context.Context, userID int64) ([]model.Subscription, error)
	CheckExpired(ctx context.Context) error
}

type subscriptionService struct {
	subRepo      repository.SubscriptionRepository
	userRepo     repository.UserRepository
	entitleSvc   EntitlementService
}

func NewSubscriptionService(
	subRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
	entitleSvc EntitlementService,
) SubscriptionService {
	return &subscriptionService{
		subRepo:    subRepo,
		userRepo:   userRepo,
		entitleSvc: entitleSvc,
	}
}

func generateOrderID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "ORD-" + hex.EncodeToString(b) + fmt.Sprintf("-%d", time.Now().UnixNano())
}

func (s *subscriptionService) PurchaseSubscription(ctx context.Context, req *model.PurchaseRequest) (*model.Subscription, error) {
	userID, err := ParseUserID(req.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	active, _ := s.subRepo.GetActiveByUserID(ctx, userID)
	if active != nil {
		return nil, ErrAlreadySubscribed
	}

	now := time.Now()
	sub := &model.Subscription{
		UserID:        userID,
		TierLevel:     req.TierLevel,
		StartTime:     now,
		EndTime:       now.Add(defaultDuration),
		Status:        "ACTIVE",
		Price:         req.Price,
		PaymentMethod: req.PaymentMethod,
		OrderID:       generateOrderID(),
	}

	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdateIdentityTier(ctx, userID, req.TierLevel); err != nil {
		return nil, err
	}

	if err := s.entitleSvc.GrantEntitlements(ctx, userID, req.TierLevel); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *subscriptionService) UpgradeSubscription(ctx context.Context, req *model.UpgradeRequest) (*model.Subscription, error) {
	userID, err := ParseUserID(req.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	active, err := s.subRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return nil, ErrSubscriptionNotFound
	}

	if req.NewTier <= active.TierLevel {
		return nil, ErrInvalidTierUpgrade
	}

	if err := s.subRepo.UpdateStatus(ctx, active.ID, "UPGRADED"); err != nil {
		return nil, err
	}

	now := time.Now()
	sub := &model.Subscription{
		UserID:        userID,
		TierLevel:     req.NewTier,
		StartTime:     now,
		EndTime:       active.EndTime,
		Status:        "ACTIVE",
		Price:         req.PriceDiff,
		PaymentMethod: req.PaymentMethod,
		OrderID:       generateOrderID(),
	}

	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdateIdentityTier(ctx, userID, req.NewTier); err != nil {
		return nil, err
	}

	if err := s.entitleSvc.GrantEntitlements(ctx, userID, req.NewTier); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *subscriptionService) RenewSubscription(ctx context.Context, req *model.RenewRequest) (*model.Subscription, error) {
	userID, err := ParseUserID(req.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	active, err := s.subRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if active == nil {
		return nil, ErrSubscriptionNotFound
	}

	newEnd := active.EndTime.Add(defaultDuration)
	if err := s.subRepo.UpdateEndTime(ctx, active.ID, newEnd.Format(time.RFC3339)); err != nil {
		return nil, err
	}

	active.EndTime = newEnd
	return active, nil
}

func (s *subscriptionService) GetUserSubscriptions(ctx context.Context, userID int64) ([]model.Subscription, error) {
	return s.subRepo.GetByUserID(ctx, userID)
}

func (s *subscriptionService) CheckExpired(ctx context.Context) error {
	subs, err := s.subRepo.FindExpired(ctx)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		if err := s.subRepo.UpdateStatus(ctx, sub.ID, "EXPIRED"); err != nil {
			continue
		}
		_ = s.userRepo.UpdateIdentityTier(ctx, sub.UserID, 0)
	}

	return nil
}
