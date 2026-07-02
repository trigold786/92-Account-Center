package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
	"github.com/trigold786/92-Account-Center/account-service/internal/repository"
	"github.com/trigold786/92-Account-Center/account-service/internal/svcconfig"
	circuitbreaker "github.com/trigold786/92-Account-Center/pkg/circuitbreaker"
)

var (
	ErrSubscriptionNotFound = errors.New("no active subscription found")
	ErrInvalidTierUpgrade   = errors.New("invalid tier upgrade")
	ErrAlreadySubscribed    = errors.New("user already has an active subscription at this or higher tier")
	ErrPaidOrderRequired    = errors.New("paid order is required before activating subscription")
)

type SubscriptionService interface {
	PurchaseSubscription(ctx context.Context, req *model.PurchaseRequest) (*model.Subscription, error)
	ActivatePaidOrderSubscription(ctx context.Context, req *model.ActivatePaidOrderRequest) (*model.Subscription, error)
	CancelRefundedOrderSubscription(ctx context.Context, req *model.CancelRefundedOrderRequest) error
	UpgradeSubscription(ctx context.Context, req *model.UpgradeRequest) (*model.Subscription, error)
	RenewSubscription(ctx context.Context, req *model.RenewRequest) (*model.Subscription, error)
	GetUserSubscriptions(ctx context.Context, userID int64) ([]model.Subscription, error)
	CheckExpired(ctx context.Context) error
}

type PaidOrder struct {
	OrderID       string
	UserID        int64
	Amount        float64
	Status        string
	PaymentMethod string
}

type PaymentOrderVerifier interface {
	VerifyPaidOrder(ctx context.Context, orderID string, userID int64, amount float64) (*PaidOrder, error)
}

type subscriptionService struct {
	subRepo         repository.SubscriptionRepository
	userRepo        repository.UserRepository
	entitleSvc      EntitlementService
	cfg             *svcconfig.AccountConfig
	paymentVerifier PaymentOrderVerifier
}

func NewSubscriptionService(
	subRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
	entitleSvc EntitlementService,
	cfg *svcconfig.AccountConfig,
) SubscriptionService {
	return NewSubscriptionServiceWithPaymentVerifier(subRepo, userRepo, entitleSvc, cfg, nil)
}

func NewSubscriptionServiceWithPaymentVerifier(
	subRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
	entitleSvc EntitlementService,
	cfg *svcconfig.AccountConfig,
	paymentVerifier PaymentOrderVerifier,
) SubscriptionService {
	return &subscriptionService{
		subRepo:         subRepo,
		userRepo:        userRepo,
		entitleSvc:      entitleSvc,
		cfg:             cfg,
		paymentVerifier: paymentVerifier,
	}
}

type HTTPPaymentOrderVerifier struct {
	baseURL string
	client  *http.Client
}

func NewHTTPPaymentOrderVerifier(baseURL string) *HTTPPaymentOrderVerifier {
	return &HTTPPaymentOrderVerifier{
		baseURL: baseURL,
		client:  circuitbreaker.WrapHTTPClient(&http.Client{Timeout: 5 * time.Second}, "payment-service"),
	}
}

func (v *HTTPPaymentOrderVerifier) VerifyPaidOrder(ctx context.Context, orderID string, userID int64, amount float64) (*PaidOrder, error) {
	if orderID == "" {
		return nil, ErrPaidOrderRequired
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/orders/%s", v.baseURL, orderID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, ErrPaidOrderRequired
	}
	var payload struct {
		Data struct {
			ID            int64   `json:"id"`
			UserID        int64   `json:"user_id"`
			Amount        float64 `json:"amount"`
			Status        string  `json:"status"`
			PaymentMethod string  `json:"payment_method"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Data.UserID != userID || payload.Data.Status != "paid" {
		return nil, ErrPaidOrderRequired
	}
	// Convert to cents for precise comparison
	expectedCents := int64(math.Round(amount * 100))
	actualCents := int64(math.Round(payload.Data.Amount * 100))
	if expectedCents != actualCents {
		return nil, ErrPaidOrderRequired
	}
	return &PaidOrder{
		OrderID:       orderID,
		UserID:        payload.Data.UserID,
		Amount:        payload.Data.Amount,
		Status:        payload.Data.Status,
		PaymentMethod: payload.Data.PaymentMethod,
	}, nil
}

func generateOrderID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
	}
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

	if req.OrderID == "" || s.paymentVerifier == nil {
		return nil, ErrPaidOrderRequired
	}
	paidOrder, err := s.paymentVerifier.VerifyPaidOrder(ctx, req.OrderID, userID, req.Price)
	if err != nil {
		return nil, err
	}
	if paidOrder == nil || paidOrder.Status != "paid" {
		return nil, ErrPaidOrderRequired
	}
	paymentMethod := req.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = paidOrder.PaymentMethod
	}

	now := time.Now()
	sub := &model.Subscription{
		UserID:        userID,
		TierLevel:     req.TierLevel,
		StartTime:     now,
		EndTime:       now.Add(s.cfg.SubscriptionDefaultDuration),
		Status:        "ACTIVE",
		Price:         req.Price,
		PaymentMethod: paymentMethod,
		OrderID:       req.OrderID,
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

func (s *subscriptionService) ActivatePaidOrderSubscription(ctx context.Context, req *model.ActivatePaidOrderRequest) (*model.Subscription, error) {
	if req.OrderID == "" || s.paymentVerifier == nil {
		return nil, ErrPaidOrderRequired
	}

	existing, err := s.subRepo.GetByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	paidOrder, err := s.paymentVerifier.VerifyPaidOrder(ctx, req.OrderID, req.UserID, req.Price)
	if err != nil {
		return nil, err
	}
	if paidOrder == nil || paidOrder.Status != "paid" || paidOrder.UserID != req.UserID {
		return nil, ErrPaidOrderRequired
	}
	// Convert to cents for precise comparison
	expectedCents := int64(req.Price * 100)
	actualCents := int64(paidOrder.Amount * 100)
	if expectedCents != actualCents {
		return nil, ErrPaidOrderRequired
	}

	paymentMethod := req.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = paidOrder.PaymentMethod
	}

	now := time.Now()
	sub := &model.Subscription{
		UserID:        req.UserID,
		TierLevel:     req.TierLevel,
		StartTime:     now,
		EndTime:       now.Add(s.cfg.SubscriptionDefaultDuration),
		Status:        "ACTIVE",
		Price:         req.Price,
		PaymentMethod: paymentMethod,
		OrderID:       req.OrderID,
	}
	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}
	if err := s.userRepo.UpdateIdentityTier(ctx, req.UserID, req.TierLevel); err != nil {
		return nil, err
	}
	if err := s.entitleSvc.GrantEntitlements(ctx, req.UserID, req.TierLevel); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *subscriptionService) CancelRefundedOrderSubscription(ctx context.Context, req *model.CancelRefundedOrderRequest) error {
	if req.OrderID == "" {
		return ErrPaidOrderRequired
	}
	sub, err := s.subRepo.GetByOrderID(ctx, req.OrderID)
	if err != nil {
		return err
	}
	if sub == nil {
		return nil
	}
	if sub.Status == "REFUNDED" || sub.Status == "CANCELLED" {
		return nil
	}
	if req.UserID != 0 && sub.UserID != req.UserID {
		return ErrPaidOrderRequired
	}
	if err := s.subRepo.UpdateStatus(ctx, sub.ID, "REFUNDED"); err != nil {
		return err
	}
	if err := s.userRepo.UpdateIdentityTier(ctx, sub.UserID, 0); err != nil {
		return err
	}
	if err := s.entitleSvc.DeleteUserEntitlements(ctx, sub.UserID); err != nil {
		return err
	}
	return nil
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

	newEnd := active.EndTime.Add(s.cfg.SubscriptionDefaultDuration)
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
		if err := s.userRepo.UpdateIdentityTier(ctx, sub.UserID, 0); err != nil {
			slog.Warn("failed to reset identity tier for expired subscription", "user_id", sub.UserID, "subscription_id", sub.ID, "error", err)
		}
	}

	return nil
}
