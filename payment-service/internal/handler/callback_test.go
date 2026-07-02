package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
	"github.com/trigold786/92-Account-Center/payment-service/internal/provider"
	"github.com/trigold786/92-Account-Center/payment-service/internal/repository"
	"github.com/trigold786/92-Account-Center/payment-service/internal/service"
)

func TestCallbackHandler_NotifiesSubscriptionActivationAfterPaidOrderUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	registry := provider.NewProviderRegistry()
	registry.Register(&callbackTestProvider{
		result: &provider.CallbackResult{
			OrderNo:       "PAY202606140002",
			TransactionID: "wx-txn-002",
			Status:        "SUCCESS",
			Amount:        99.9,
		},
	})

	orderRepo := &callbackTestOrderRepo{order: &model.Order{
		ID:          102,
		OrderNo:     "PAY202606140002",
		UserID:      7,
		ProductType: "subscription",
		Amount:      99.9,
		Status:      model.OrderStatusPending,
		Metadata:    `{"tier_level":2}`,
	}}
	orderSvc := &callbackTestOrderService{updatedOrder: &model.Order{ID: 102, Status: model.OrderStatusPaid}}
	callbackRepo := &callbackTestRepository{}
	notifier := &callbackTestActivationNotifier{}

	h := NewCallbackHandlerWithActivationNotifier(registry, orderRepo, orderSvc, callbackRepo, notifier, slog.Default())
	r := gin.New()
	r.POST("/callback/:provider", h.HandleCallback)

	body := `{"order_no":"PAY202606140002","transaction_id":"wx-txn-002","status":"SUCCESS"}`
	req := httptest.NewRequest(http.MethodPost, "/callback/wechat", strings.NewReader(body))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if notifier.request == nil {
		t.Fatal("expected subscription activation notifier to be called")
	}
	if notifier.request.UserID != 7 {
		t.Fatalf("expected user 7, got %d", notifier.request.UserID)
	}
	if notifier.request.OrderID != "102" {
		t.Fatalf("expected order id 102, got %q", notifier.request.OrderID)
	}
	if notifier.request.TierLevel != 2 {
		t.Fatalf("expected tier 2, got %d", notifier.request.TierLevel)
	}
	if notifier.request.Price != 99.9 {
		t.Fatalf("expected price 99.9, got %f", notifier.request.Price)
	}
	if notifier.request.PaymentMethod != "wechat" {
		t.Fatalf("expected payment method wechat, got %q", notifier.request.PaymentMethod)
	}
}

func TestCallbackHandler_PersistsVerifiedCallbackBeforeOrderUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	registry := provider.NewProviderRegistry()
	registry.Register(&callbackTestProvider{
		result: &provider.CallbackResult{
			OrderNo:       "PAY202606140001",
			TransactionID: "wx-txn-001",
			Status:        "SUCCESS",
			Amount:        99.9,
		},
	})

	orderRepo := &callbackTestOrderRepo{order: &model.Order{
		ID:      101,
		OrderNo: "PAY202606140001",
		UserID:  1,
		Amount:  99.9,
		Status:  model.OrderStatusPending,
	}}
	orderSvc := &callbackTestOrderService{updatedOrder: &model.Order{ID: 101, Status: model.OrderStatusPaid}}
	callbackRepo := &callbackTestRepository{}

	h := NewCallbackHandlerWithCallbackRepository(registry, orderRepo, orderSvc, callbackRepo, slog.Default())
	r := gin.New()
	r.POST("/callback/:provider", h.HandleCallback)

	body := `{"order_no":"PAY202606140001","transaction_id":"wx-txn-001","status":"SUCCESS"}`
	req := httptest.NewRequest(http.MethodPost, "/callback/wechat", strings.NewReader(body))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if callbackRepo.saved == nil {
		t.Fatal("expected verified callback to be persisted")
	}
	if callbackRepo.saved.Provider != "wechat" {
		t.Fatalf("expected provider wechat, got %q", callbackRepo.saved.Provider)
	}
	if callbackRepo.saved.OrderNo != "PAY202606140001" {
		t.Fatalf("expected order no persisted, got %q", callbackRepo.saved.OrderNo)
	}
	if callbackRepo.saved.TransactionID != "wx-txn-001" {
		t.Fatalf("expected transaction id persisted, got %q", callbackRepo.saved.TransactionID)
	}
	if callbackRepo.saved.Status != "SUCCESS" {
		t.Fatalf("expected status SUCCESS, got %q", callbackRepo.saved.Status)
	}
	if !callbackRepo.saved.Verified {
		t.Fatal("expected persisted callback to be marked verified")
	}
	if callbackRepo.saved.RawPayload != body {
		t.Fatalf("expected raw payload %q, got %q", body, callbackRepo.saved.RawPayload)
	}
}

type callbackTestActivationNotifier struct {
	request *SubscriptionActivationRequest
}

func (n *callbackTestActivationNotifier) NotifySubscriptionActivation(ctx context.Context, req SubscriptionActivationRequest) error {
	n.request = &req
	return nil
}

type callbackTestProvider struct {
	result *provider.CallbackResult
	err    error
}

func (p *callbackTestProvider) Name() string { return "wechat" }
func (p *callbackTestProvider) CreatePayment(context.Context, *provider.CreatePaymentRequest) (*provider.CreatePaymentResponse, error) {
	return nil, errors.New("not implemented")
}
func (p *callbackTestProvider) QueryPayment(context.Context, string) (*provider.PaymentStatus, error) {
	return nil, errors.New("not implemented")
}
func (p *callbackTestProvider) Refund(context.Context, *provider.RefundRequest) (*provider.RefundResponse, error) {
	return nil, errors.New("not implemented")
}
func (p *callbackTestProvider) VerifyCallback(context.Context, map[string]string, []byte) (*provider.CallbackResult, error) {
	return p.result, p.err
}

type callbackTestRepository struct {
	saved *model.PaymentCallback
}

func (r *callbackTestRepository) Save(ctx context.Context, callback *model.PaymentCallback) error {
	r.saved = callback
	return nil
}

type callbackTestOrderRepo struct {
	repository.OrderRepository
	order *model.Order
}

func (r *callbackTestOrderRepo) GetByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	return r.order, nil
}

type callbackTestOrderService struct {
	service.OrderService
	updatedOrder *model.Order
}

func (s *callbackTestOrderService) UpdateStatus(ctx context.Context, id int64, req *model.UpdateOrderStatusRequest) (*model.Order, error) {
	return s.updatedOrder, nil
}
