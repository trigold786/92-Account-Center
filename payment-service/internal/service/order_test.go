package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
	"github.com/trigold786/92-Account-Center/payment-service/internal/svcconfig"
)

type mockOrderRepo struct {
	mock.Mock
}

func (m *mockOrderRepo) Create(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *mockOrderRepo) GetByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	args := m.Called(ctx, orderNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *mockOrderRepo) List(ctx context.Context, query *model.OrderQueryRequest) ([]model.Order, int, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]model.Order), args.Int(1), args.Error(2)
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id int64, status model.OrderStatus, paymentMethod, paymentTxnID string) error {
	args := m.Called(ctx, id, status, paymentMethod, paymentTxnID)
	return args.Error(0)
}

func (m *mockOrderRepo) UpdateRefund(ctx context.Context, id int64, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *mockOrderRepo) FindExpired(ctx context.Context, before time.Time) ([]model.Order, error) {
	args := m.Called(ctx, before)
	return args.Get(0).([]model.Order), args.Error(1)
}

func (m *mockOrderRepo) GetPendingOrdersOlderThan(ctx context.Context, since time.Duration) ([]*model.Order, error) {
	args := m.Called(ctx, since)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Order), args.Error(1)
}

func (m *mockOrderRepo) UpdateOrderStatus(ctx context.Context, orderNo string, fromStatus, toStatus string) error {
	args := m.Called(ctx, orderNo, fromStatus, toStatus)
	return args.Error(0)
}

func defaultPaymentConfig() *svcconfig.PaymentConfig {
	return &svcconfig.PaymentConfig{
		DefaultPageSize:    20,
		OrderExpiryMinutes: 30,
	}
}

func ptrOrder(o model.Order) *model.Order {
	return &o
}

func TestCreateOrder_Success(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	req := &model.CreateOrderRequest{
		UserID:      1,
		ProductType: "subscription",
		ProductName: "premium",
		Amount:      99.99,
		Currency:    "CNY",
	}

	repo.On("Create", mock.Anything, mock.AnythingOfType("*model.Order")).Return(nil)

	order, err := svc.CreateOrder(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), order.UserID)
	assert.Equal(t, model.OrderStatusPending, order.Status)
	assert.Equal(t, "subscription", order.ProductType)
	repo.AssertExpectations(t)
}

func TestCreateOrder_RepoError(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	req := &model.CreateOrderRequest{
		UserID:      1,
		ProductType: "subscription",
		ProductName: "premium",
		Amount:      99.99,
	}

	repo.On("Create", mock.Anything, mock.AnythingOfType("*model.Order")).Return(errors.New("db error"))

	order, err := svc.CreateOrder(context.Background(), req)
	assert.Nil(t, order)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestGetOrder_Success(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	now := time.Now()
	expected := &model.Order{ID: 1, OrderNo: "PAY001", UserID: 1, Status: model.OrderStatusPending, CreatedAt: now, UpdatedAt: now}
	repo.On("GetByID", mock.Anything, int64(1)).Return(expected, nil)

	order, err := svc.GetOrder(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), order.ID)
	assert.Equal(t, "PAY001", order.OrderNo)
	repo.AssertExpectations(t)
}

func TestGetOrder_NotFound(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	repo.On("GetByID", mock.Anything, int64(99)).Return(nil, nil)

	order, err := svc.GetOrder(context.Background(), 99)
	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrOrderNotFound)
	repo.AssertExpectations(t)
}

func TestGetOrderByNo_Success(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	now := time.Now()
	expected := &model.Order{ID: 1, OrderNo: "PAY001", UserID: 1, Status: model.OrderStatusPending, CreatedAt: now, UpdatedAt: now}
	repo.On("GetByOrderNo", mock.Anything, "PAY001").Return(expected, nil)

	order, err := svc.GetOrderByNo(context.Background(), "PAY001")
	assert.NoError(t, err)
	assert.Equal(t, "PAY001", order.OrderNo)
	repo.AssertExpectations(t)
}

func TestGetOrderByNo_NotFound(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	repo.On("GetByOrderNo", mock.Anything, "PAY999").Return(nil, nil)

	order, err := svc.GetOrderByNo(context.Background(), "PAY999")
	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrOrderNotFound)
	repo.AssertExpectations(t)
}

func TestListOrders_Success(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	orders := []model.Order{
		{ID: 1, OrderNo: "PAY001", UserID: 1},
		{ID: 2, OrderNo: "PAY002", UserID: 2},
	}

	repo.On("List", mock.Anything, mock.AnythingOfType("*model.OrderQueryRequest")).Return(orders, 2, nil)

	query := &model.OrderQueryRequest{Page: 1, PageSize: 20}
	resp, err := svc.ListOrders(context.Background(), query)
	assert.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Orders, 2)
	repo.AssertExpectations(t)
}

func TestListOrders_DefaultPagination(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	repo.On("List", mock.Anything, mock.AnythingOfType("*model.OrderQueryRequest")).Return([]model.Order{}, 0, nil)

	query := &model.OrderQueryRequest{}
	resp, err := svc.ListOrders(context.Background(), query)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 20, resp.PageSize)
	repo.AssertExpectations(t)
}

func TestListOrders_EmptyResult(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	repo.On("List", mock.Anything, mock.AnythingOfType("*model.OrderQueryRequest")).Return([]model.Order(nil), 0, nil)

	query := &model.OrderQueryRequest{Page: 1, PageSize: 10}
	resp, err := svc.ListOrders(context.Background(), query)
	assert.NoError(t, err)
	assert.NotNil(t, resp.Orders)
	assert.Empty(t, resp.Orders)
	repo.AssertExpectations(t)
}

func TestListOrders_RepoError(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	repo.On("List", mock.Anything, mock.AnythingOfType("*model.OrderQueryRequest")).Return([]model.Order(nil), 0, errors.New("db error"))

	query := &model.OrderQueryRequest{Page: 1, PageSize: 10}
	resp, err := svc.ListOrders(context.Background(), query)
	assert.Nil(t, resp)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestUpdateStatus_PendingToPaid(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	now := time.Now()
	existing := &model.Order{ID: 1, Status: model.OrderStatusPending, CreatedAt: now, UpdatedAt: now}
	updated := &model.Order{ID: 1, Status: model.OrderStatusPaid, CreatedAt: now, UpdatedAt: now}

	repo.On("GetByID", mock.Anything, int64(1)).Return(existing, nil).Once()
	repo.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusPaid, "wechat", "txn123").Return(nil)
	repo.On("GetByID", mock.Anything, int64(1)).Return(updated, nil).Once()

	req := &model.UpdateOrderStatusRequest{
		Status:              model.OrderStatusPaid,
		PaymentMethod:       "wechat",
		PaymentTransactionID: "txn123",
	}
	order, err := svc.UpdateStatus(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.Equal(t, model.OrderStatusPaid, order.Status)
	repo.AssertExpectations(t)
}

func TestUpdateStatus_PendingToCancelled(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	now := time.Now()
	existing := &model.Order{ID: 1, Status: model.OrderStatusPending, CreatedAt: now, UpdatedAt: now}
	updated := &model.Order{ID: 1, Status: model.OrderStatusCancelled, CreatedAt: now, UpdatedAt: now}

	repo.On("GetByID", mock.Anything, int64(1)).Return(existing, nil).Once()
	repo.On("UpdateStatus", mock.Anything, int64(1), model.OrderStatusCancelled, "", "").Return(nil)
	repo.On("GetByID", mock.Anything, int64(1)).Return(updated, nil).Once()

	req := &model.UpdateOrderStatusRequest{Status: model.OrderStatusCancelled}
	order, err := svc.UpdateStatus(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.Equal(t, model.OrderStatusCancelled, order.Status)
	repo.AssertExpectations(t)
}

func TestUpdateStatus_PaidToRefunded(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	now := time.Now()
	existing := &model.Order{ID: 1, Status: model.OrderStatusPaid, CreatedAt: now, UpdatedAt: now}
	updated := &model.Order{ID: 1, Status: model.OrderStatusRefunded, CreatedAt: now, UpdatedAt: now}

	repo.On("GetByID", mock.Anything, int64(1)).Return(existing, nil).Once()
	repo.On("UpdateRefund", mock.Anything, int64(1), "defective").Return(nil)
	repo.On("GetByID", mock.Anything, int64(1)).Return(updated, nil).Once()

	req := &model.UpdateOrderStatusRequest{
		Status:       model.OrderStatusRefunded,
		RefundReason: "defective",
	}
	order, err := svc.UpdateStatus(context.Background(), 1, req)
	assert.NoError(t, err)
	assert.Equal(t, model.OrderStatusRefunded, order.Status)
	repo.AssertExpectations(t)
}

func TestUpdateStatus_InvalidTransition_PaidToPending(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	now := time.Now()
	existing := &model.Order{ID: 1, Status: model.OrderStatusPaid, CreatedAt: now, UpdatedAt: now}
	repo.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)

	req := &model.UpdateOrderStatusRequest{Status: model.OrderStatusPending}
	order, err := svc.UpdateStatus(context.Background(), 1, req)
	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrInvalidTransition)
	repo.AssertExpectations(t)
}

func TestUpdateStatus_InvalidTransition_RefundedToPaid(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	now := time.Now()
	existing := &model.Order{ID: 1, Status: model.OrderStatusRefunded, CreatedAt: now, UpdatedAt: now}
	repo.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)

	req := &model.UpdateOrderStatusRequest{Status: model.OrderStatusPaid}
	order, err := svc.UpdateStatus(context.Background(), 1, req)
	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrInvalidTransition)
	repo.AssertExpectations(t)
}

func TestUpdateStatus_InvalidTransition_CancelledToPaid(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	now := time.Now()
	existing := &model.Order{ID: 1, Status: model.OrderStatusCancelled, CreatedAt: now, UpdatedAt: now}
	repo.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)

	req := &model.UpdateOrderStatusRequest{Status: model.OrderStatusPaid}
	order, err := svc.UpdateStatus(context.Background(), 1, req)
	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrInvalidTransition)
	repo.AssertExpectations(t)
}

func TestUpdateStatus_InvalidTransition_PaidToCancelled(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	now := time.Now()
	existing := &model.Order{ID: 1, Status: model.OrderStatusPaid, CreatedAt: now, UpdatedAt: now}
	repo.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)

	req := &model.UpdateOrderStatusRequest{Status: model.OrderStatusCancelled}
	order, err := svc.UpdateStatus(context.Background(), 1, req)
	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrInvalidTransition)
	repo.AssertExpectations(t)
}

func TestUpdateStatus_OrderNotFound(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	repo.On("GetByID", mock.Anything, int64(99)).Return(nil, nil)

	req := &model.UpdateOrderStatusRequest{Status: model.OrderStatusPaid}
	order, err := svc.UpdateStatus(context.Background(), 99, req)
	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrOrderNotFound)
	repo.AssertExpectations(t)
}

func TestUpdateStatus_GetByIDError(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	repo.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

	req := &model.UpdateOrderStatusRequest{Status: model.OrderStatusPaid}
	order, err := svc.UpdateStatus(context.Background(), 1, req)
	assert.Nil(t, order)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestExportCSV_Success(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	now := time.Now()
	orders := []model.Order{
		{ID: 1, OrderNo: "PAY001", UserID: 1, ProductType: "sub", ProductName: "premium", Amount: 99.99, Currency: "CNY", Status: model.OrderStatusPaid, CreatedAt: now, UpdatedAt: now},
	}

	repo.On("List", mock.Anything, mock.AnythingOfType("*model.OrderQueryRequest")).Return(orders, 1, nil)

	query := &model.OrderQueryRequest{Page: 1, PageSize: 20}
	data, err := svc.ExportCSV(context.Background(), query)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	reader := csv.NewReader(bytes.NewReader(data))
	records, err := reader.ReadAll()
	assert.NoError(t, err)
	assert.Len(t, records, 2)
	assert.Equal(t, "id", records[0][0])
	assert.Equal(t, "PAY001", records[1][1])
	assert.Equal(t, "99.99", records[1][5])
	repo.AssertExpectations(t)
}

func TestExportCSV_Empty(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	repo.On("List", mock.Anything, mock.AnythingOfType("*model.OrderQueryRequest")).Return([]model.Order{}, 0, nil)

	query := &model.OrderQueryRequest{Page: 1, PageSize: 20}
	data, err := svc.ExportCSV(context.Background(), query)
	assert.NoError(t, err)

	reader := csv.NewReader(bytes.NewReader(data))
	records, err := reader.ReadAll()
	assert.NoError(t, err)
	assert.Len(t, records, 1)
	assert.Equal(t, "id", records[0][0])
	repo.AssertExpectations(t)
}

func TestExportCSV_RepoError(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	repo.On("List", mock.Anything, mock.AnythingOfType("*model.OrderQueryRequest")).Return([]model.Order(nil), 0, errors.New("db error"))

	query := &model.OrderQueryRequest{Page: 1, PageSize: 20}
	data, err := svc.ExportCSV(context.Background(), query)
	assert.Nil(t, data)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestExportCSV_HeaderFormat(t *testing.T) {
	repo := new(mockOrderRepo)
	cfg := defaultPaymentConfig()
	svc := NewOrderService(repo, cfg).(*orderService)

	repo.On("List", mock.Anything, mock.AnythingOfType("*model.OrderQueryRequest")).Return([]model.Order{}, 0, nil)

	query := &model.OrderQueryRequest{Page: 1, PageSize: 20}
	data, err := svc.ExportCSV(context.Background(), query)
	assert.NoError(t, err)

	reader := csv.NewReader(bytes.NewReader(data))
	record, err := reader.Read()
	assert.NoError(t, err)
	assert.Equal(t, []string{"id", "order_no", "user_id", "product_type", "product_name", "amount", "currency", "status", "payment_method", "payment_transaction_id", "created_at", "updated_at"}, record)

	_, err = reader.Read()
	assert.Equal(t, io.EOF, err)
	repo.AssertExpectations(t)
}

func TestIsValidTransition_AllValid(t *testing.T) {
	assert.True(t, isValidTransition(model.OrderStatusPending, model.OrderStatusPaid))
	assert.True(t, isValidTransition(model.OrderStatusPending, model.OrderStatusCancelled))
	assert.True(t, isValidTransition(model.OrderStatusPaid, model.OrderStatusRefunded))
}

func TestIsValidTransition_AllInvalid(t *testing.T) {
	assert.False(t, isValidTransition(model.OrderStatusPaid, model.OrderStatusPending))
	assert.False(t, isValidTransition(model.OrderStatusCancelled, model.OrderStatusPaid))
	assert.False(t, isValidTransition(model.OrderStatusRefunded, model.OrderStatusPaid))
	assert.False(t, isValidTransition(model.OrderStatusPaid, model.OrderStatusCancelled))
	assert.False(t, isValidTransition(model.OrderStatusRefunded, model.OrderStatusCancelled))
	assert.False(t, isValidTransition(model.OrderStatusCancelled, model.OrderStatusRefunded))
	assert.False(t, isValidTransition(model.OrderStatusPending, model.OrderStatusPending))
	assert.False(t, isValidTransition(model.OrderStatusPaid, model.OrderStatusPaid))
}
