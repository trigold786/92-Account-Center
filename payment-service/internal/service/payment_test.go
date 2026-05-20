package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
	"github.com/trigold786/92-Account-Center/payment-service/internal/provider"
)

func TestProviderRegistry_RegisterAndGet(t *testing.T) {
	registry := provider.NewProviderRegistry()
	p := &mockProvider{name: "test"}
	registry.Register(p)

	got, ok := registry.Get("test")
	assert.True(t, ok)
	assert.Equal(t, "test", got.Name())
}

func TestProviderRegistry_GetNotFound(t *testing.T) {
	registry := provider.NewProviderRegistry()
	_, ok := registry.Get("nonexistent")
	assert.False(t, ok)
}

func TestProviderRegistry_List(t *testing.T) {
	registry := provider.NewProviderRegistry()
	registry.Register(&mockProvider{name: "wechat"})
	registry.Register(&mockProvider{name: "alipay"})

	names := registry.List()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "wechat")
	assert.Contains(t, names, "alipay")
}

func TestWeChatProvider_Name(t *testing.T) {
	p := NewWeChatPayProvider(WeChatPayConfig{
		AppID:  "wx_test",
		MchID:  "mch_test",
		APIKey: "key_test",
	})
	assert.Equal(t, "wechat", p.Name())
}

func TestWeChatProvider_CreatePayment_H5(t *testing.T) {
	p := NewWeChatPayProvider(WeChatPayConfig{
		AppID:  "wx_test",
		MchID:  "mch_test",
		APIKey: "key_test",
	})
	req := &provider.CreatePaymentRequest{
		OrderNo:   "PAY001",
		Amount:    100.00,
		Currency:  "CNY",
		Subject:   "test",
		TradeType: "wechat_h5",
		NotifyURL: "http://localhost/callback",
	}
	resp, err := p.CreatePayment(context.Background(), req)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.PaymentURL)
	assert.NotEmpty(t, resp.PrepayID)
	assert.NotEmpty(t, resp.TransactionID)
	assert.Empty(t, resp.QRCodeURL)
}

func TestWeChatProvider_CreatePayment_Native(t *testing.T) {
	p := NewWeChatPayProvider(WeChatPayConfig{
		AppID:  "wx_test",
		MchID:  "mch_test",
		APIKey: "key_test",
	})
	req := &provider.CreatePaymentRequest{
		OrderNo:   "PAY001",
		Amount:    100.00,
		TradeType: "wechat_native",
		NotifyURL: "http://localhost/callback",
	}
	resp, err := p.CreatePayment(context.Background(), req)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.QRCodeURL)
}

func TestWeChatProvider_CreatePayment_Mini(t *testing.T) {
	p := NewWeChatPayProvider(WeChatPayConfig{
		AppID:  "wx_test",
		MchID:  "mch_test",
		APIKey: "key_test",
	})
	req := &provider.CreatePaymentRequest{
		OrderNo:   "PAY002",
		Amount:    50.00,
		TradeType: "wechat_mini",
		NotifyURL: "http://localhost/callback",
	}
	resp, err := p.CreatePayment(context.Background(), req)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.PaymentURL)
}

func TestWeChatProvider_CreatePayment_InvalidTradeType(t *testing.T) {
	p := NewWeChatPayProvider(WeChatPayConfig{})
	req := &provider.CreatePaymentRequest{
		TradeType: "alipay_wap",
	}
	resp, err := p.CreatePayment(context.Background(), req)
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestWeChatProvider_QueryPayment(t *testing.T) {
	p := NewWeChatPayProvider(WeChatPayConfig{})
	status, err := p.QueryPayment(context.Background(), "PAY001")
	assert.NoError(t, err)
	assert.Equal(t, "PAY001", status.OrderNo)
	assert.Equal(t, "NOTPAY", status.Status)
}

func TestWeChatProvider_Refund(t *testing.T) {
	p := NewWeChatPayProvider(WeChatPayConfig{})
	req := &provider.RefundRequest{
		OrderNo:      "PAY001",
		RefundNo:     "REF001",
		TotalAmount:  100.00,
		RefundAmount: 50.00,
		Reason:       "test",
	}
	resp, err := p.Refund(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "REF001", resp.RefundNo)
	assert.Equal(t, "SUCCESS", resp.Status)
}

func TestWeChatProvider_VerifyCallback(t *testing.T) {
	apiKey := "test_key"
	p := NewWeChatPayProvider(WeChatPayConfig{APIKey: apiKey})
	body, _ := json.Marshal(map[string]interface{}{
		"out_trade_no":   "PAY001",
		"transaction_id": "TXN001",
		"result_code":    "SUCCESS",
		"total_amount":   "100.00",
		"time_paid":      "2025-01-01 12:00:00",
	})
	timestamp := "1234567890"
	nonce := "abc123"
	message := timestamp + "\n" + nonce + "\n" + string(body) + "\n"
	mac := hmac.New(sha256.New, []byte(apiKey))
	mac.Write([]byte(message))
	sig := hex.EncodeToString(mac.Sum(nil))

	headers := map[string]string{
		"Wechatpay-Timestamp": timestamp,
		"Wechatpay-Nonce":     nonce,
		"Wechatpay-Signature": sig,
	}
	result, err := p.VerifyCallback(context.Background(), headers, body)
	assert.NoError(t, err)
	assert.Equal(t, "PAY001", result.OrderNo)
	assert.Equal(t, "TXN001", result.TransactionID)
	assert.Equal(t, "SUCCESS", result.Status)
}

func TestAlipayProvider_Name(t *testing.T) {
	p := NewAlipayProvider(AlipayConfig{
		AppID: "alipay_test",
	})
	assert.Equal(t, "alipay", p.Name())
}

func TestAlipayProvider_CreatePayment_WAP(t *testing.T) {
	p := NewAlipayProvider(AlipayConfig{AppID: "alipay_test"})
	req := &provider.CreatePaymentRequest{
		OrderNo:   "PAY001",
		Amount:    200.00,
		TradeType: "alipay_wap",
		NotifyURL: "http://localhost/callback",
	}
	resp, err := p.CreatePayment(context.Background(), req)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.PaymentURL)
	assert.NotEmpty(t, resp.PrepayID)
}

func TestAlipayProvider_CreatePayment_App(t *testing.T) {
	p := NewAlipayProvider(AlipayConfig{AppID: "alipay_test"})
	req := &provider.CreatePaymentRequest{
		OrderNo:   "PAY002",
		Amount:    300.00,
		TradeType: "alipay_app",
		NotifyURL: "http://localhost/callback",
	}
	resp, err := p.CreatePayment(context.Background(), req)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.PaymentURL)
}

func TestAlipayProvider_CreatePayment_InvalidTradeType(t *testing.T) {
	p := NewAlipayProvider(AlipayConfig{})
	req := &provider.CreatePaymentRequest{
		TradeType: "wechat_h5",
	}
	resp, err := p.CreatePayment(context.Background(), req)
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestAlipayProvider_QueryPayment(t *testing.T) {
	p := NewAlipayProvider(AlipayConfig{})
	status, err := p.QueryPayment(context.Background(), "PAY001")
	assert.NoError(t, err)
	assert.Equal(t, "PAY001", status.OrderNo)
}

func TestAlipayProvider_Refund(t *testing.T) {
	p := NewAlipayProvider(AlipayConfig{})
	req := &provider.RefundRequest{
		RefundNo:     "REF001",
		TotalAmount:  200.00,
		RefundAmount: 200.00,
	}
	resp, err := p.Refund(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "REF001", resp.RefundNo)
	assert.Equal(t, "REFUND_SUCCESS", resp.Status)
}

func TestAlipayProvider_VerifyCallback_Success(t *testing.T) {
	p := NewAlipayProvider(AlipayConfig{PrivateKey: "test_key"})
	body, _ := json.Marshal(map[string]interface{}{
		"out_trade_no": "PAY001",
		"trade_no":     "ALITXN001",
		"trade_status": "TRADE_SUCCESS",
		"total_amount": "200.00",
		"gmt_payment":  "2025-01-01 12:00:00",
	})
	result, err := p.VerifyCallback(context.Background(), map[string]string{}, body)
	assert.NoError(t, err)
	assert.Equal(t, "PAY001", result.OrderNo)
	assert.Equal(t, "SUCCESS", result.Status)
}

func TestAlipayProvider_VerifyCallback_Fail(t *testing.T) {
	p := NewAlipayProvider(AlipayConfig{PrivateKey: "test_key"})
	body, _ := json.Marshal(map[string]interface{}{
		"out_trade_no": "PAY001",
		"trade_status": "TRADE_CLOSED",
	})
	result, err := p.VerifyCallback(context.Background(), map[string]string{}, body)
	assert.NoError(t, err)
	assert.Equal(t, "FAIL", result.Status)
}

type mockReconOrderRepo struct {
	mock.Mock
}

func (m *mockReconOrderRepo) Create(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockReconOrderRepo) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *mockReconOrderRepo) GetByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	args := m.Called(ctx, orderNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *mockReconOrderRepo) List(ctx context.Context, query *model.OrderQueryRequest) ([]model.Order, int, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]model.Order), args.Int(1), args.Error(2)
}

func (m *mockReconOrderRepo) UpdateStatus(ctx context.Context, id int64, status model.OrderStatus, paymentMethod, paymentTxnID string) error {
	args := m.Called(ctx, id, status, paymentMethod, paymentTxnID)
	return args.Error(0)
}

func (m *mockReconOrderRepo) UpdateRefund(ctx context.Context, id int64, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func TestReconciliation_MismatchDetection(t *testing.T) {
	registry := provider.NewProviderRegistry()
	mockP := &mockQueryProvider{
		status: &provider.PaymentStatus{
			OrderNo:       "PAY001",
			TransactionID: "TXN001",
			Status:        "CLOSED",
			Amount:        50.00,
		},
	}
	registry.Register(mockP)

	repo := new(mockReconOrderRepo)
	now := time.Now()
	orders := []model.Order{
		{ID: 1, OrderNo: "PAY001", Amount: 100.00, Status: model.OrderStatusPaid, CreatedAt: now, UpdatedAt: now},
	}
	repo.On("List", mock.Anything, mock.AnythingOfType("*model.OrderQueryRequest")).Return(orders, 1, nil)

	svc := NewReconciliationService(registry, repo, nil).(*reconciliationService)

	report, err := svc.ReconcileOrders(context.Background(), "mock", now.Format("2006-01-02"))
	assert.NoError(t, err)
	assert.Equal(t, 1, report.TotalOrders)
	assert.Equal(t, 0, report.MatchedOrders)
	assert.Len(t, report.MismatchOrders, 1)
	assert.Equal(t, "PAY001", report.MismatchOrders[0].OrderNo)
	assert.Equal(t, "paid", report.MismatchOrders[0].LocalStatus)
	assert.Equal(t, "CLOSED", report.MismatchOrders[0].RemoteStatus)
	repo.AssertExpectations(t)
}

func TestReconciliation_MatchedOrders(t *testing.T) {
	registry := provider.NewProviderRegistry()
	mockP := &mockQueryProvider{
		status: &provider.PaymentStatus{
			OrderNo:       "PAY001",
			TransactionID: "TXN001",
			Status:        "SUCCESS",
			Amount:        100.00,
		},
	}
	registry.Register(mockP)

	repo := new(mockReconOrderRepo)
	now := time.Now()
	orders := []model.Order{
		{ID: 1, OrderNo: "PAY001", Amount: 100.00, Status: model.OrderStatusPaid, CreatedAt: now, UpdatedAt: now},
	}
	repo.On("List", mock.Anything, mock.AnythingOfType("*model.OrderQueryRequest")).Return(orders, 1, nil)

	svc := NewReconciliationService(registry, repo, nil).(*reconciliationService)

	report, err := svc.ReconcileOrders(context.Background(), "mock", now.Format("2006-01-02"))
	assert.NoError(t, err)
	assert.Equal(t, 1, report.TotalOrders)
	assert.Equal(t, 1, report.MatchedOrders)
	assert.Empty(t, report.MismatchOrders)
	repo.AssertExpectations(t)
}

func TestReconciliation_GetReport(t *testing.T) {
	registry := provider.NewProviderRegistry()
	mockP := &mockQueryProvider{
		status: &provider.PaymentStatus{Status: "SUCCESS", Amount: 100.00},
	}
	registry.Register(mockP)

	repo := new(mockReconOrderRepo)
	now := time.Now()
	orders := []model.Order{
		{ID: 1, OrderNo: "PAY001", Amount: 100.00, Status: model.OrderStatusPaid, CreatedAt: now, UpdatedAt: now},
	}
	repo.On("List", mock.Anything, mock.AnythingOfType("*model.OrderQueryRequest")).Return(orders, 1, nil)

	svc := NewReconciliationService(registry, repo, nil).(*reconciliationService)

	dateStr := now.Format("2006-01-02")
	report, err := svc.ReconcileOrders(context.Background(), "mock", dateStr)
	assert.NoError(t, err)

	stored, err := svc.GetReconciliationReport(context.Background(), report.ID)
	assert.NoError(t, err)
	assert.Equal(t, report.ID, stored.ID)
}

func TestReconciliation_GetReportNotFound(t *testing.T) {
	registry := provider.NewProviderRegistry()
	repo := new(mockReconOrderRepo)
	svc := NewReconciliationService(registry, repo, nil).(*reconciliationService)

	_, err := svc.GetReconciliationReport(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestReconciliation_UnknownProvider(t *testing.T) {
	registry := provider.NewProviderRegistry()
	repo := new(mockReconOrderRepo)
	svc := NewReconciliationService(registry, repo, nil).(*reconciliationService)

	_, err := svc.ReconcileOrders(context.Background(), "unknown", "2025-01-01")
	assert.Error(t, err)
}

func TestReconciliation_InvalidDate(t *testing.T) {
	registry := provider.NewProviderRegistry()
	registry.Register(&mockQueryProvider{})
	repo := new(mockReconOrderRepo)
	svc := NewReconciliationService(registry, repo, nil).(*reconciliationService)

	_, err := svc.ReconcileOrders(context.Background(), "mock", "invalid-date")
	assert.Error(t, err)
}

func TestCallbackIdempotency_OrderAlreadyPaid(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	now := time.Now()
	paidOrder := &model.Order{
		ID:                   1,
		OrderNo:              "PAY001",
		Status:               model.OrderStatusPaid,
		PaymentMethod:        "wechat",
		PaymentTransactionID: "TXN001",
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	repo.On("GetByID", mock.Anything, int64(1)).Return(paidOrder, nil)

	req := &model.UpdateOrderStatusRequest{
		Status:               model.OrderStatusPaid,
		PaymentMethod:        "wechat",
		PaymentTransactionID: "TXN002",
	}
	order, err := svc.UpdateStatus(context.Background(), 1, req)
	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrInvalidTransition)
	repo.AssertExpectations(t)
}

func TestStateMachine_PaymentSuccessUpdatesOrder(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	now := time.Now()
	pendingOrder := &model.Order{
		ID:        1,
		OrderNo:   "PAY001",
		Status:    model.OrderStatusPending,
		Amount:    100.00,
		CreatedAt: now,
		UpdatedAt: now,
	}
	paidOrder := &model.Order{
		ID:                   1,
		OrderNo:              "PAY001",
		Status:               model.OrderStatusPaid,
		PaymentMethod:        "wechat",
		PaymentTransactionID: "TXN001",
		Amount:               100.00,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	repo.On("GetByID", mock.Anything, int64(1)).Return(pendingOrder, nil).Once()
	repo.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusPaid, "wechat", "TXN001").Return(nil)
	repo.On("GetByID", mock.Anything, int64(1)).Return(paidOrder, nil).Once()

	req := &model.UpdateOrderStatusRequest{
		Status:               model.OrderStatusPaid,
		PaymentMethod:        "wechat",
		PaymentTransactionID: "TXN001",
	}
	order, err := svc.UpdateStatus(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.Equal(t, model.OrderStatusPaid, order.Status)
	assert.Equal(t, "wechat", order.PaymentMethod)
	assert.Equal(t, "TXN001", order.PaymentTransactionID)
	repo.AssertExpectations(t)
}

type mockProvider struct {
	name string
}

func (m *mockProvider) CreatePayment(ctx context.Context, req *provider.CreatePaymentRequest) (*provider.CreatePaymentResponse, error) {
	return nil, nil
}
func (m *mockProvider) QueryPayment(ctx context.Context, orderNo string) (*provider.PaymentStatus, error) {
	return nil, nil
}
func (m *mockProvider) Refund(ctx context.Context, req *provider.RefundRequest) (*provider.RefundResponse, error) {
	return nil, nil
}
func (m *mockProvider) VerifyCallback(ctx context.Context, headers map[string]string, body []byte) (*provider.CallbackResult, error) {
	return nil, nil
}
func (m *mockProvider) Name() string { return m.name }

type mockQueryProvider struct {
	status *provider.PaymentStatus
}

func (m *mockQueryProvider) CreatePayment(ctx context.Context, req *provider.CreatePaymentRequest) (*provider.CreatePaymentResponse, error) {
	return &provider.CreatePaymentResponse{}, nil
}
func (m *mockQueryProvider) QueryPayment(ctx context.Context, orderNo string) (*provider.PaymentStatus, error) {
	if m.status != nil {
		m.status.OrderNo = orderNo
		return m.status, nil
	}
	return &provider.PaymentStatus{OrderNo: orderNo, Status: "SUCCESS"}, nil
}
func (m *mockQueryProvider) Refund(ctx context.Context, req *provider.RefundRequest) (*provider.RefundResponse, error) {
	return &provider.RefundResponse{}, nil
}
func (m *mockQueryProvider) VerifyCallback(ctx context.Context, headers map[string]string, body []byte) (*provider.CallbackResult, error) {
	return &provider.CallbackResult{}, nil
}
func (m *mockQueryProvider) Name() string { return "mock" }

type mockReconRepo struct {
	mock.Mock
}

func (m *mockReconRepo) Create(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}
func (m *mockReconRepo) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}
func (m *mockReconRepo) GetByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	args := m.Called(ctx, orderNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}
func (m *mockReconRepo) List(ctx context.Context, query *model.OrderQueryRequest) ([]model.Order, int, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]model.Order), args.Int(1), args.Error(2)
}
func (m *mockReconRepo) UpdateStatus(ctx context.Context, id int64, status model.OrderStatus, paymentMethod, paymentTxnID string) error {
	args := m.Called(ctx, id, status, paymentMethod, paymentTxnID)
	return args.Error(0)
}
func (m *mockReconRepo) UpdateRefund(ctx context.Context, id int64, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func TestMapLocalStatus(t *testing.T) {
	assert.Equal(t, "SUCCESS", mapLocalStatus(model.OrderStatusPaid))
	assert.Equal(t, "NOTPAY", mapLocalStatus(model.OrderStatusPending))
	assert.Equal(t, "CLOSED", mapLocalStatus(model.OrderStatusCancelled))
	assert.Equal(t, "REFUND", mapLocalStatus(model.OrderStatusRefunded))
}

func TestResolveProvider(t *testing.T) {
	// imported from handler package - tested via handler tests
}

func TestReconciliation_QueryError_PartialStatus(t *testing.T) {
	registry := provider.NewProviderRegistry()
	registry.Register(&errorQueryProvider{})
	repo := new(mockReconOrderRepo)
	now := time.Now()
	orders := []model.Order{
		{ID: 1, OrderNo: "PAY001", Amount: 100.00, Status: model.OrderStatusPaid, CreatedAt: now, UpdatedAt: now},
	}
	repo.On("List", mock.Anything, mock.AnythingOfType("*model.OrderQueryRequest")).Return(orders, 1, nil)

	testLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	svc := NewReconciliationService(registry, repo, testLogger).(*reconciliationService)
	report, err := svc.ReconcileOrders(context.Background(), "errprov", now.Format("2006-01-02"))
	assert.NoError(t, err)
	assert.Equal(t, "partial", report.Status)
}

type errorQueryProvider struct{}

func (e *errorQueryProvider) CreatePayment(ctx context.Context, req *provider.CreatePaymentRequest) (*provider.CreatePaymentResponse, error) {
	return nil, nil
}
func (e *errorQueryProvider) QueryPayment(ctx context.Context, orderNo string) (*provider.PaymentStatus, error) {
	return nil, errors.New("network timeout")
}
func (e *errorQueryProvider) Refund(ctx context.Context, req *provider.RefundRequest) (*provider.RefundResponse, error) {
	return nil, nil
}
func (e *errorQueryProvider) VerifyCallback(ctx context.Context, headers map[string]string, body []byte) (*provider.CallbackResult, error) {
	return nil, nil
}
func (e *errorQueryProvider) Name() string { return "errprov" }
