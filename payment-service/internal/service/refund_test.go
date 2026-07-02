package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
	"github.com/trigold786/92-Account-Center/payment-service/internal/provider"
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

func TestRequestRefundRejectsNonPaidOrder(t *testing.T) {
	repo := &mockRefundRepo{}
	orderRepo := &mockRefundOrderRepo{order: &model.Order{
		ID:     1,
		Status: model.OrderStatusPending,
	}}
	svc := NewRefundService(repo, orderRepo, nil)
	_, err := svc.RequestRefund(context.Background(), 1, 42, "test")
	if err != ErrOrderNotPaid {
		t.Fatalf("expected ErrOrderNotPaid, got %v", err)
	}
}

func TestRefundService_ApproveRefund_CancelsSubscriptionOrder(t *testing.T) {
	repo := &mockRefundRepo{}
	orderRepo := &mockRefundOrderRepo{order: &model.Order{
		ID:          102,
		UserID:      42,
		Status:      model.OrderStatusPaid,
		ProductType: "subscription",
		Amount:      100,
		CreatedAt:   time.Now().Add(-2 * 24 * time.Hour),
	}}
	notifier := &mockSubscriptionCancellationNotifier{}
	svc := NewRefundService(repo, orderRepo, nil, notifier)

	err := svc.ApproveRefund(context.Background(), 1, 9)
	if err != nil {
		t.Fatalf("ApproveRefund failed: %v", err)
	}
	if !orderRepo.refunded {
		t.Fatal("expected order to be marked refunded")
	}
	if !notifier.called {
		t.Fatal("expected subscription cancellation notifier to be called")
	}
	if notifier.userID != 42 || notifier.orderID != 102 || notifier.reason != "refund approved" {
		t.Fatalf("unexpected cancellation payload: user=%d order=%d reason=%q", notifier.userID, notifier.orderID, notifier.reason)
	}
}

func TestRefundService_ApproveRefund_CallsProviderBeforeInternalReversal(t *testing.T) {
	order := &model.Order{
		ID:            102,
		UserID:        42,
		OrderNo:       "PAY102",
		Status:        model.OrderStatusPaid,
		ProductType:   "subscription",
		Amount:        100,
		PaymentMethod: "wechat_native",
		CreatedAt:     time.Now().Add(-2 * 24 * time.Hour),
	}
	orderRepo := &mockRefundOrderRepo{order: order}
	notifier := &mockSubscriptionCancellationNotifier{}
	pp := &mockPaymentProvider{name: "wechat"}
	registry := &mockProviderRegistry{providers: map[string]provider.PaymentProvider{"wechat": pp}}
	repo := &mockRefundRepo{refund: &model.Refund{ID: 5, Status: "pending", OrderID: 102, Amount: 100, UserID: 42}}
	svc := NewRefundService(repo, orderRepo, nil, notifier, registry)

	err := svc.ApproveRefund(context.Background(), 5, 9)
	if err != nil {
		t.Fatalf("ApproveRefund failed: %v", err)
	}
	if !pp.called {
		t.Fatal("expected provider refund to be called")
	}
	if pp.lastReq.OrderNo != "PAY102" {
		t.Fatalf("expected provider request order no PAY102, got %q", pp.lastReq.OrderNo)
	}
	if pp.lastReq.TotalAmount != 100 {
		t.Fatalf("expected provider request total amount 100, got %v", pp.lastReq.TotalAmount)
	}
	if pp.lastReq.RefundAmount != 100 {
		t.Fatalf("expected provider request refund amount 100, got %v", pp.lastReq.RefundAmount)
	}
	if repo.providerRefundID != "WXREFUND123" {
		t.Fatalf("expected persisted provider refund id WXREFUND123, got %q", repo.providerRefundID)
	}
	if repo.providerStatus != "SUCCESS" {
		t.Fatalf("expected persisted provider status SUCCESS, got %q", repo.providerStatus)
	}
	if repo.approvedStatus != "approved" {
		t.Fatalf("expected refund to become approved, got %q", repo.approvedStatus)
	}
	if !orderRepo.refunded {
		t.Fatal("expected order to be marked refunded")
	}
	if !notifier.called {
		t.Fatal("expected subscription cancellation notifier to be called")
	}
}

func TestRefundService_ApproveRefund_ProviderFailureBlocksInternalReversal(t *testing.T) {
	order := &model.Order{
		ID:            103,
		UserID:        42,
		OrderNo:       "PAY103",
		Status:        model.OrderStatusPaid,
		ProductType:   "subscription",
		Amount:        100,
		PaymentMethod: "alipay_wap",
		CreatedAt:     time.Now().Add(-2 * 24 * time.Hour),
	}
	orderRepo := &mockRefundOrderRepo{order: order}
	notifier := &mockSubscriptionCancellationNotifier{}
	pp := &mockPaymentProvider{name: "alipay", err: errors.New("provider down")}
	registry := &mockProviderRegistry{providers: map[string]provider.PaymentProvider{"alipay": pp}}
	repo := &mockRefundRepo{refund: &model.Refund{ID: 6, Status: "pending", OrderID: 103, Amount: 100, UserID: 42}}
	svc := NewRefundService(repo, orderRepo, nil, notifier, registry)

	err := svc.ApproveRefund(context.Background(), 6, 9)
	if err == nil {
		t.Fatal("expected error from ApproveRefund when provider fails, got nil")
	}
	if repo.failedStatus != "failed" {
		t.Fatalf("expected refund to be marked failed, got %q", repo.failedStatus)
	}
	if repo.providerError == "" {
		t.Fatal("expected non-empty provider error to be persisted")
	}
	if orderRepo.refunded {
		t.Fatal("expected order NOT to be marked refunded on provider failure")
	}
	if notifier.called {
		t.Fatal("expected subscription cancellation NOT to be called on provider failure")
	}
}

func TestRefundService_ApproveRefund_UnknownPaymentMethodFailsBeforeReversal(t *testing.T) {
	order := &model.Order{
		ID:            104,
		UserID:        42,
		OrderNo:       "PAY104",
		Status:        model.OrderStatusPaid,
		ProductType:   "subscription",
		Amount:        100,
		PaymentMethod: "bank_transfer",
		CreatedAt:     time.Now().Add(-2 * 24 * time.Hour),
	}
	orderRepo := &mockRefundOrderRepo{order: order}
	notifier := &mockSubscriptionCancellationNotifier{}
	registry := &mockProviderRegistry{providers: map[string]provider.PaymentProvider{}}
	repo := &mockRefundRepo{refund: &model.Refund{ID: 7, Status: "pending", OrderID: 104, Amount: 100, UserID: 42}}
	svc := NewRefundService(repo, orderRepo, nil, notifier, registry)

	err := svc.ApproveRefund(context.Background(), 7, 9)
	if err == nil {
		t.Fatal("expected error from ApproveRefund for unknown payment method, got nil")
	}
	if repo.failedStatus != "failed" {
		t.Fatalf("expected refund to be marked failed, got %q", repo.failedStatus)
	}
	if repo.providerError == "" {
		t.Fatal("expected non-empty provider error to be persisted")
	}
	if orderRepo.refunded {
		t.Fatal("expected order NOT to be marked refunded for unknown payment method")
	}
	if notifier.called {
		t.Fatal("expected subscription cancellation NOT to be called for unknown payment method")
	}
}

func TestRefundStatusFlow(t *testing.T) {
	repo := &mockRefundRepo{}
	orderRepo := &mockRefundOrderRepo{order: &model.Order{
		ID:        1,
		UserID:    42,
		Status:    model.OrderStatusPaid,
		Amount:    100,
		CreatedAt: time.Now().Add(-2 * 24 * time.Hour),
	}}
	svc := NewRefundService(repo, orderRepo, nil)
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

type mockSubscriptionCancellationNotifier struct {
	called  bool
	userID  int64
	orderID int64
	reason  string
}

func (m *mockSubscriptionCancellationNotifier) CancelRefundedOrderSubscription(ctx context.Context, userID int64, orderID int64, reason string) error {
	m.called = true
	m.userID = userID
	m.orderID = orderID
	m.reason = reason
	return nil
}

type mockRefundOrderRepo struct {
	order    *model.Order
	refunded bool
}

func (m *mockRefundOrderRepo) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	if m.order != nil {
		return m.order, nil
	}
	return nil, nil
}

func (m *mockRefundOrderRepo) GetPendingOrdersOlderThan(ctx context.Context, since time.Duration) ([]*model.Order, error) {
	return nil, nil
}

func (m *mockRefundOrderRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	if id == m.order.ID && status == string(model.OrderStatusRefunded) {
		m.refunded = true
		m.order.Status = model.OrderStatusRefunded
	}
	return nil
}

func (m *mockRefundOrderRepo) UpdateOrderStatus(ctx context.Context, orderNo string, fromStatus, toStatus string) error {
	return nil
}

type mockRefundRepo struct {
	refund           *model.Refund
	providerRefundID string
	providerStatus   string
	providerError    string
	failedStatus     string
	approvedStatus   string
}

func (m *mockRefundRepo) FindByOrderID(ctx context.Context, orderID int64) (*model.Refund, error) {
	return nil, nil
}

func (m *mockRefundRepo) Create(ctx context.Context, r *model.Refund) error {
	r.ID = 1
	return nil
}

func (m *mockRefundRepo) GetByID(ctx context.Context, id int64) (*model.Refund, error) {
	if m.refund != nil {
		return m.refund, nil
	}
	return &model.Refund{ID: id, Status: "pending", OrderID: 1, Amount: 100, UserID: 42}, nil
}

func (m *mockRefundRepo) UpdateStatus(ctx context.Context, id int64, status string, approverID int64, note string) error {
	m.approvedStatus = status
	return nil
}

func (m *mockRefundRepo) UpdateProviderResult(ctx context.Context, id int64, refundNo string, providerName string, providerRefundID string, providerStatus string) error {
	m.providerRefundID = providerRefundID
	m.providerStatus = providerStatus
	return nil
}

func (m *mockRefundRepo) MarkProviderFailure(ctx context.Context, id int64, refundNo string, providerName string, providerError string) error {
	m.providerError = providerError
	m.failedStatus = "failed"
	return nil
}

type mockPaymentProvider struct {
	name    string
	err     error
	called  bool
	lastReq *provider.RefundRequest
}

func (m *mockPaymentProvider) CreatePayment(ctx context.Context, req *provider.CreatePaymentRequest) (*provider.CreatePaymentResponse, error) {
	return nil, nil
}

func (m *mockPaymentProvider) QueryPayment(ctx context.Context, orderNo string) (*provider.PaymentStatus, error) {
	return nil, nil
}

func (m *mockPaymentProvider) Refund(ctx context.Context, req *provider.RefundRequest) (*provider.RefundResponse, error) {
	m.called = true
	m.lastReq = req
	if m.err != nil {
		return nil, m.err
	}
	return &provider.RefundResponse{RefundNo: req.RefundNo, Status: "SUCCESS", RefundID: "WXREFUND123"}, nil
}

func (m *mockPaymentProvider) VerifyCallback(ctx context.Context, headers map[string]string, body []byte) (*provider.CallbackResult, error) {
	return nil, nil
}

func (m *mockPaymentProvider) Name() string {
	return m.name
}

type mockProviderRegistry struct {
	providers map[string]provider.PaymentProvider
}

func (m *mockProviderRegistry) Get(name string) (provider.PaymentProvider, bool) {
	p, ok := m.providers[name]
	return p, ok
}
