package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
	"github.com/trigold786/92-Account-Center/account-service/internal/repository"
	"github.com/trigold786/92-Account-Center/account-service/internal/svcconfig"
)

type mockSubscriptionRepo struct {
	mock.Mock
}

func (m *mockSubscriptionRepo) Create(ctx context.Context, sub *model.Subscription) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *mockSubscriptionRepo) GetActiveByUserID(ctx context.Context, userID int64) (*model.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepo) GetByUserID(ctx context.Context, userID int64) ([]model.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepo) GetByOrderID(ctx context.Context, orderID string) (*model.Subscription, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	return m.Called(ctx, id, status).Error(0)
}

func (m *mockSubscriptionRepo) UpdateEndTime(ctx context.Context, id int64, endTime string) error {
	return m.Called(ctx, id, endTime).Error(0)
}

func (m *mockSubscriptionRepo) FindExpired(ctx context.Context) ([]model.Subscription, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Subscription), args.Error(1)
}

type mockEntitlementService struct {
	mock.Mock
}

type mockPaymentVerifier struct {
	mock.Mock
}

func (m *mockPaymentVerifier) VerifyPaidOrder(ctx context.Context, orderID string, userID int64, amount float64) (*PaidOrder, error) {
	args := m.Called(ctx, orderID, userID, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PaidOrder), args.Error(1)
}

func (m *mockEntitlementService) GetUserEntitlements(ctx context.Context, userID int64) ([]model.Entitlement, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Entitlement), args.Error(1)
}

func (m *mockEntitlementService) ConsumeQuota(ctx context.Context, userID int64, featureCode string, amount int) (*model.ConsumeResponse, error) {
	args := m.Called(ctx, userID, featureCode, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ConsumeResponse), args.Error(1)
}

func (m *mockEntitlementService) GrantEntitlements(ctx context.Context, userID int64, tierLevel int) error {
	return m.Called(ctx, userID, tierLevel).Error(0)
}

func (m *mockEntitlementService) DeleteUserEntitlements(ctx context.Context, userID int64) error {
	return m.Called(ctx, userID).Error(0)
}

var subTestCfg = &svcconfig.AccountConfig{
	DeletionFreezeDays:          7,
	SubscriptionDefaultDuration: 720 * time.Hour,
	EntitlementCacheTTL:         24 * time.Hour,
}

func newTestSubService(subRepo repository.SubscriptionRepository, userRepo repository.UserRepository, entSvc EntitlementService) SubscriptionService {
	return NewSubscriptionService(subRepo, userRepo, entSvc, subTestCfg)
}

func newTestSubServiceWithVerifier(subRepo repository.SubscriptionRepository, userRepo repository.UserRepository, entSvc EntitlementService, verifier PaymentOrderVerifier) SubscriptionService {
	return NewSubscriptionServiceWithPaymentVerifier(subRepo, userRepo, entSvc, subTestCfg, verifier)
}

func newPaidOrderVerifier(orderID string, userID int64, amount float64, paymentMethod string) *mockPaymentVerifier {
	verifier := new(mockPaymentVerifier)
	verifier.On("VerifyPaidOrder", mock.Anything, orderID, userID, amount).Return(&PaidOrder{
		OrderID:       orderID,
		UserID:        userID,
		Amount:        amount,
		Status:        "paid",
		PaymentMethod: paymentMethod,
	}, nil)
	return verifier
}

func TestSubscriptionService_Interface(t *testing.T) {
	var _ repository.SubscriptionRepository = (*mockSubscriptionRepo)(nil)
	var _ EntitlementService = (*mockEntitlementService)(nil)
}

func TestSubscriptionService_Purchase_Success(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(nil, nil)
	subRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Subscription")).Return(nil)
	userRepo.On("UpdateIdentityTier", mock.Anything, int64(1), 2).Return(nil)
	entSvc.On("GrantEntitlements", mock.Anything, int64(1), 2).Return(nil)
	verifier := newPaidOrderVerifier("1001", 1, 99.9, "alipay")

	svc := newTestSubServiceWithVerifier(subRepo, userRepo, entSvc, verifier)
	sub, err := svc.PurchaseSubscription(context.Background(), &model.PurchaseRequest{
		UserID:        "1",
		TierLevel:     2,
		Price:         99.9,
		PaymentMethod: "alipay",
		OrderID:       "1001",
	})

	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, int64(1), sub.UserID)
	assert.Equal(t, 2, sub.TierLevel)
	assert.Equal(t, "ACTIVE", sub.Status)
	assert.Equal(t, 99.9, sub.Price)
	assert.Equal(t, "1001", sub.OrderID)
	subRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	entSvc.AssertExpectations(t)
}

func TestSubscriptionService_Purchase_RequiresPaidOrder(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(nil, nil)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.PurchaseSubscription(context.Background(), &model.PurchaseRequest{
		UserID:        "1",
		TierLevel:     2,
		Price:         99.9,
		PaymentMethod: "alipay",
	})

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "paid order")
	}
	assert.Nil(t, sub)
	subRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "UpdateIdentityTier", mock.Anything, mock.Anything, mock.Anything)
	entSvc.AssertNotCalled(t, "GrantEntitlements", mock.Anything, mock.Anything, mock.Anything)
}

func TestSubscriptionService_ActivatePaidOrder_IsIdempotentByOrderID(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	existing := &model.Subscription{ID: 88, UserID: 1, TierLevel: 2, Status: "ACTIVE", OrderID: "1001"}
	subRepo.On("GetByOrderID", mock.Anything, "1001").Return(existing, nil)

	svc := newTestSubServiceWithVerifier(subRepo, userRepo, entSvc, new(mockPaymentVerifier))
	sub, err := svc.ActivatePaidOrderSubscription(context.Background(), &model.ActivatePaidOrderRequest{
		UserID:        1,
		OrderID:       "1001",
		TierLevel:     2,
		Price:         99.9,
		PaymentMethod: "wechat",
	})

	assert.NoError(t, err)
	assert.Equal(t, existing, sub)
	subRepo.AssertExpectations(t)
	subRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "UpdateIdentityTier", mock.Anything, mock.Anything, mock.Anything)
	entSvc.AssertNotCalled(t, "GrantEntitlements", mock.Anything, mock.Anything, mock.Anything)
}

func TestSubscriptionService_CancelRefundedOrder_CancelsSubscriptionAndDeletesEntitlements(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	sub := &model.Subscription{ID: 55, UserID: 7, TierLevel: 2, Status: "ACTIVE", OrderID: "102"}
	subRepo.On("GetByOrderID", mock.Anything, "102").Return(sub, nil)
	subRepo.On("UpdateStatus", mock.Anything, int64(55), "REFUNDED").Return(nil)
	userRepo.On("UpdateIdentityTier", mock.Anything, int64(7), 0).Return(nil)
	entSvc.On("DeleteUserEntitlements", mock.Anything, int64(7)).Return(nil)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	err := svc.CancelRefundedOrderSubscription(context.Background(), &model.CancelRefundedOrderRequest{
		UserID:  7,
		OrderID: "102",
		Reason:  "refund approved",
	})

	assert.NoError(t, err)
	subRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	entSvc.AssertExpectations(t)
}

func TestSubscriptionService_Purchase_InvalidUserID(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.PurchaseSubscription(context.Background(), &model.PurchaseRequest{
		UserID:    "abc",
		TierLevel: 2,
		Price:     50.0,
	})

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, sub)
}

func TestSubscriptionService_Purchase_AlreadySubscribed(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	activeSub := &model.Subscription{ID: 10, UserID: 1, TierLevel: 2, Status: "ACTIVE"}
	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(activeSub, nil)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.PurchaseSubscription(context.Background(), &model.PurchaseRequest{
		UserID:    "1",
		TierLevel: 3,
		Price:     100.0,
	})

	assert.Error(t, err)
	assert.Equal(t, ErrAlreadySubscribed, err)
	assert.Nil(t, sub)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_Purchase_CreateError(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(nil, nil)
	subRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Subscription")).Return(errors.New("db error"))
	verifier := newPaidOrderVerifier("1002", 1, 99.0, "alipay")

	svc := newTestSubServiceWithVerifier(subRepo, userRepo, entSvc, verifier)
	sub, err := svc.PurchaseSubscription(context.Background(), &model.PurchaseRequest{
		UserID:    "1",
		TierLevel: 2,
		Price:     99.0,
		OrderID:   "1002",
	})

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.Nil(t, sub)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_Purchase_UpdateTierError(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(nil, nil)
	subRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Subscription")).Return(nil)
	userRepo.On("UpdateIdentityTier", mock.Anything, int64(1), 2).Return(errors.New("tier update failed"))
	verifier := newPaidOrderVerifier("1003", 1, 99.0, "wechat")

	svc := newTestSubServiceWithVerifier(subRepo, userRepo, entSvc, verifier)
	sub, err := svc.PurchaseSubscription(context.Background(), &model.PurchaseRequest{
		UserID:    "1",
		TierLevel: 2,
		Price:     99.0,
		OrderID:   "1003",
	})

	assert.Error(t, err)
	assert.Equal(t, "tier update failed", err.Error())
	assert.Nil(t, sub)
	subRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestSubscriptionService_Purchase_GrantEntitlementsError(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(nil, nil)
	subRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Subscription")).Return(nil)
	userRepo.On("UpdateIdentityTier", mock.Anything, int64(1), 2).Return(nil)
	entSvc.On("GrantEntitlements", mock.Anything, int64(1), 2).Return(errors.New("grant failed"))
	verifier := newPaidOrderVerifier("1004", 1, 99.0, "wechat")

	svc := newTestSubServiceWithVerifier(subRepo, userRepo, entSvc, verifier)
	sub, err := svc.PurchaseSubscription(context.Background(), &model.PurchaseRequest{
		UserID:    "1",
		TierLevel: 2,
		Price:     99.0,
		OrderID:   "1004",
	})

	assert.Error(t, err)
	assert.Equal(t, "grant failed", err.Error())
	assert.Nil(t, sub)
	subRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	entSvc.AssertExpectations(t)
}

func TestSubscriptionService_Upgrade_Success(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	activeSub := &model.Subscription{
		ID:        10,
		UserID:    1,
		TierLevel: 2,
		EndTime:   time.Now().Add(600 * time.Hour),
		Status:    "ACTIVE",
	}
	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(activeSub, nil)
	subRepo.On("UpdateStatus", mock.Anything, int64(10), "UPGRADED").Return(nil)
	subRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Subscription")).Return(nil)
	userRepo.On("UpdateIdentityTier", mock.Anything, int64(1), 3).Return(nil)
	entSvc.On("GrantEntitlements", mock.Anything, int64(1), 3).Return(nil)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.UpgradeSubscription(context.Background(), &model.UpgradeRequest{
		UserID:        "1",
		NewTier:       3,
		PriceDiff:     50.0,
		PaymentMethod: "wechat",
	})

	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, int64(1), sub.UserID)
	assert.Equal(t, 3, sub.TierLevel)
	assert.Equal(t, "ACTIVE", sub.Status)
	assert.Equal(t, 50.0, sub.Price)
	assert.Equal(t, activeSub.EndTime.Unix(), sub.EndTime.Unix())
	subRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	entSvc.AssertExpectations(t)
}

func TestSubscriptionService_Upgrade_InvalidUserID(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.UpgradeSubscription(context.Background(), &model.UpgradeRequest{
		UserID:  "invalid",
		NewTier: 3,
	})

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, sub)
}

func TestSubscriptionService_Upgrade_NoActiveSubscription(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(nil, nil)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.UpgradeSubscription(context.Background(), &model.UpgradeRequest{
		UserID:  "1",
		NewTier: 3,
	})

	assert.Error(t, err)
	assert.Equal(t, ErrSubscriptionNotFound, err)
	assert.Nil(t, sub)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_Upgrade_GetActiveError(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.UpgradeSubscription(context.Background(), &model.UpgradeRequest{
		UserID:  "1",
		NewTier: 3,
	})

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.Nil(t, sub)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_Upgrade_InvalidTierDowngrade(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	activeSub := &model.Subscription{ID: 10, UserID: 1, TierLevel: 3, Status: "ACTIVE"}
	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(activeSub, nil)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.UpgradeSubscription(context.Background(), &model.UpgradeRequest{
		UserID:  "1",
		NewTier: 2,
	})

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidTierUpgrade, err)
	assert.Nil(t, sub)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_Upgrade_SameTier(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	activeSub := &model.Subscription{ID: 10, UserID: 1, TierLevel: 3, Status: "ACTIVE"}
	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(activeSub, nil)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.UpgradeSubscription(context.Background(), &model.UpgradeRequest{
		UserID:  "1",
		NewTier: 3,
	})

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidTierUpgrade, err)
	assert.Nil(t, sub)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_Upgrade_UpdateStatusError(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	activeSub := &model.Subscription{ID: 10, UserID: 1, TierLevel: 2, Status: "ACTIVE"}
	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(activeSub, nil)
	subRepo.On("UpdateStatus", mock.Anything, int64(10), "UPGRADED").Return(errors.New("status error"))

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.UpgradeSubscription(context.Background(), &model.UpgradeRequest{
		UserID:  "1",
		NewTier: 3,
	})

	assert.Error(t, err)
	assert.Equal(t, "status error", err.Error())
	assert.Nil(t, sub)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_Upgrade_CreateError(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	activeSub := &model.Subscription{ID: 10, UserID: 1, TierLevel: 2, EndTime: time.Now().Add(600 * time.Hour), Status: "ACTIVE"}
	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(activeSub, nil)
	subRepo.On("UpdateStatus", mock.Anything, int64(10), "UPGRADED").Return(nil)
	subRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Subscription")).Return(errors.New("create error"))

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.UpgradeSubscription(context.Background(), &model.UpgradeRequest{
		UserID:  "1",
		NewTier: 3,
	})

	assert.Error(t, err)
	assert.Equal(t, "create error", err.Error())
	assert.Nil(t, sub)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_Renew_Success(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	endTime := time.Now().Add(48 * time.Hour)
	activeSub := &model.Subscription{
		ID:        10,
		UserID:    1,
		TierLevel: 2,
		EndTime:   endTime,
		Status:    "ACTIVE",
	}
	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(activeSub, nil)
	subRepo.On("UpdateEndTime", mock.Anything, int64(10), mock.AnythingOfType("string")).Return(nil)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.RenewSubscription(context.Background(), &model.RenewRequest{
		UserID:        "1",
		PaymentMethod: "alipay",
	})

	assert.NoError(t, err)
	assert.NotNil(t, sub)
	expectedEnd := endTime.Add(subTestCfg.SubscriptionDefaultDuration)
	assert.Equal(t, expectedEnd.Unix(), sub.EndTime.Unix())
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_Renew_InvalidUserID(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.RenewSubscription(context.Background(), &model.RenewRequest{
		UserID: "xyz",
	})

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, sub)
}

func TestSubscriptionService_Renew_NoActiveSubscription(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(nil, nil)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.RenewSubscription(context.Background(), &model.RenewRequest{
		UserID: "1",
	})

	assert.Error(t, err)
	assert.Equal(t, ErrSubscriptionNotFound, err)
	assert.Nil(t, sub)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_Renew_GetActiveError(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.RenewSubscription(context.Background(), &model.RenewRequest{
		UserID: "1",
	})

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.Nil(t, sub)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_Renew_UpdateEndTimeError(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	activeSub := &model.Subscription{
		ID:        10,
		UserID:    1,
		TierLevel: 2,
		EndTime:   time.Now().Add(48 * time.Hour),
		Status:    "ACTIVE",
	}
	subRepo.On("GetActiveByUserID", mock.Anything, int64(1)).Return(activeSub, nil)
	subRepo.On("UpdateEndTime", mock.Anything, int64(10), mock.AnythingOfType("string")).Return(errors.New("update error"))

	svc := newTestSubService(subRepo, userRepo, entSvc)
	sub, err := svc.RenewSubscription(context.Background(), &model.RenewRequest{
		UserID: "1",
	})

	assert.Error(t, err)
	assert.Equal(t, "update error", err.Error())
	assert.Nil(t, sub)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_GetUserSubscriptions_Success(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	subs := []model.Subscription{
		{ID: 1, UserID: 1, TierLevel: 2, Status: "EXPIRED"},
		{ID: 2, UserID: 1, TierLevel: 3, Status: "ACTIVE"},
	}
	subRepo.On("GetByUserID", mock.Anything, int64(1)).Return(subs, nil)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	result, err := svc.GetUserSubscriptions(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, int64(2), result[1].ID)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_GetUserSubscriptions_Empty(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	subRepo.On("GetByUserID", mock.Anything, int64(99)).Return(nil, nil)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	result, err := svc.GetUserSubscriptions(context.Background(), 99)

	assert.NoError(t, err)
	assert.Nil(t, result)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_GetUserSubscriptions_Error(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	subRepo.On("GetByUserID", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

	svc := newTestSubService(subRepo, userRepo, entSvc)
	result, err := svc.GetUserSubscriptions(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_CheckExpired_NoExpired(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	subRepo.On("FindExpired", mock.Anything).Return(nil, nil)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	err := svc.CheckExpired(context.Background())

	assert.NoError(t, err)
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_CheckExpired_WithExpired(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	expired := []model.Subscription{
		{ID: 1, UserID: 10, Status: "ACTIVE"},
		{ID: 2, UserID: 20, Status: "ACTIVE"},
	}
	subRepo.On("FindExpired", mock.Anything).Return(expired, nil)
	subRepo.On("UpdateStatus", mock.Anything, int64(1), "EXPIRED").Return(nil)
	subRepo.On("UpdateStatus", mock.Anything, int64(2), "EXPIRED").Return(nil)
	userRepo.On("UpdateIdentityTier", mock.Anything, int64(10), 0).Return(nil)
	userRepo.On("UpdateIdentityTier", mock.Anything, int64(20), 0).Return(nil)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	err := svc.CheckExpired(context.Background())

	assert.NoError(t, err)
	subRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestSubscriptionService_CheckExpired_FindError(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	subRepo.On("FindExpired", mock.Anything).Return(nil, errors.New("query error"))

	svc := newTestSubService(subRepo, userRepo, entSvc)
	err := svc.CheckExpired(context.Background())

	assert.Error(t, err)
	assert.Equal(t, "query error", err.Error())
	subRepo.AssertExpectations(t)
}

func TestSubscriptionService_CheckExpired_PartialUpdateStatusFailure(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	expired := []model.Subscription{
		{ID: 1, UserID: 10, Status: "ACTIVE"},
		{ID: 2, UserID: 20, Status: "ACTIVE"},
	}
	subRepo.On("FindExpired", mock.Anything).Return(expired, nil)
	subRepo.On("UpdateStatus", mock.Anything, int64(1), "EXPIRED").Return(errors.New("status update failed"))
	subRepo.On("UpdateStatus", mock.Anything, int64(2), "EXPIRED").Return(nil)
	userRepo.On("UpdateIdentityTier", mock.Anything, int64(20), 0).Return(nil)

	svc := newTestSubService(subRepo, userRepo, entSvc)
	err := svc.CheckExpired(context.Background())

	assert.NoError(t, err)
	subRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestSubscriptionService_Purchase_SetsEndTimeFromConfig(t *testing.T) {
	subRepo := new(mockSubscriptionRepo)
	userRepo := new(mockUserRepo)
	entSvc := new(mockEntitlementService)

	var capturedSub *model.Subscription
	subRepo.On("GetActiveByUserID", mock.Anything, int64(5)).Return(nil, nil)
	subRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *model.Subscription) bool {
		capturedSub = s
		return true
	})).Return(nil)
	userRepo.On("UpdateIdentityTier", mock.Anything, int64(5), 2).Return(nil)
	entSvc.On("GrantEntitlements", mock.Anything, int64(5), 2).Return(nil)
	verifier := newPaidOrderVerifier("1005", 5, 99.0, "alipay")

	svc := newTestSubServiceWithVerifier(subRepo, userRepo, entSvc, verifier)
	_, err := svc.PurchaseSubscription(context.Background(), &model.PurchaseRequest{
		UserID:    "5",
		TierLevel: 2,
		Price:     99.0,
		OrderID:   "1005",
	})

	assert.NoError(t, err)
	assert.NotNil(t, capturedSub)
	assert.WithinDuration(t, capturedSub.StartTime.Add(subTestCfg.SubscriptionDefaultDuration), capturedSub.EndTime, time.Second)
	subRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	entSvc.AssertExpectations(t)
}
